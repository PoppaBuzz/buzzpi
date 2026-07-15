package bpp

import (
	"testing"
)

func TestNewChannel(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)
	if ch == nil {
		t.Fatal("NewChannel returned nil")
	}
	if ch.ID() != ChannelRPC {
		t.Errorf("ID() = %d, want %d", ch.ID(), ChannelRPC)
	}
	if ch.Mode() != ModeReliableOrdered {
		t.Errorf("Mode() = %v, want %v", ch.Mode(), ModeReliableOrdered)
	}
}

func TestChannelSend(t *testing.T) {
	t.Parallel()

	var sent *Packet
	sendFn := func(pkt *Packet) error {
		sent = pkt
		return nil
	}
	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)
	err := ch.Send([]byte("hello"))
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if sent == nil {
		t.Fatal("packet not sent")
	}
	if string(sent.Payload) != "hello" {
		t.Errorf("Payload = %q, want %q", string(sent.Payload), "hello")
	}
}

func TestChannelReceive(t *testing.T) {
	t.Parallel()

	var sentAck *Packet
	sendFn := func(pkt *Packet) error {
		if pkt.Header.PType == PacketTypeAck {
			sentAck = pkt
		}
		return nil
	}
	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)
	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("data"))
	pkt.Header.Sequence = 0
	err := ch.Receive(pkt)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if sentAck == nil {
		t.Error("ACK not sent")
	}
}

func TestChannelRead(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelStream, ModeStream, sendFn)

	pkt := NewPacket(PacketTypeRequest, ChannelStream, []byte("stream data"))
	_ = ch.Receive(pkt)

	buf := make([]byte, 100)
	n, err := ch.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "stream data" {
		t.Errorf("Read = %q, want %q", string(buf[:n]), "stream data")
	}
}

func TestChannelReadNotStream(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)
	buf := make([]byte, 10)
	_, err := ch.Read(buf)
	if err == nil {
		t.Fatal("expected error for non-stream channel Read")
	}
}

func TestChannelClose(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelStream, ModeStream, sendFn)
	err := ch.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestChannelStats(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)
	_ = ch.Send([]byte("test"))
	stats := ch.Stats()
	if stats.PacketsSent != 1 {
		t.Errorf("PacketsSent = %d, want 1", stats.PacketsSent)
	}
}

func TestChannelModeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode ChannelMode
		want string
	}{
		{ModeReliableOrdered, "reliable_ordered"},
		{ModeReliableUnordered, "reliable_unordered"},
		{ModeUnreliableUnordered, "unreliable_unordered"},
		{ModeStream, "stream"},
		{ChannelMode(99), "unknown"},
	}
	for _, tt := range tests {
		got := tt.mode.String()
		if got != tt.want {
			t.Errorf("ChannelMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestDefaultChannelConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   ChannelID
		mode ChannelMode
	}{
		{ChannelControl, ModeReliableOrdered},
		{ChannelRPC, ModeReliableOrdered},
		{ChannelEvent, ModeReliableOrdered},
		{ChannelStream, ModeStream},
		{ChannelReliable, ModeReliableUnordered},
		{ChannelUnreliable, ModeUnreliableUnordered},
	}
	for _, tt := range tests {
		cfg := DefaultChannelConfig(tt.id)
		if cfg.Mode != tt.mode {
			t.Errorf("DefaultChannelConfig(%d).Mode = %v, want %v", tt.id, cfg.Mode, tt.mode)
		}
	}
}

func TestMultiplexer(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	mux := NewMultiplexer(sendFn)

	ch := mux.Open(ChannelRPC, ModeReliableOrdered)
	if ch == nil {
		t.Fatal("Open returned nil")
	}

	got := mux.Get(ChannelRPC)
	if got != ch {
		t.Error("Get returned different channel")
	}

	if mux.Get(ChannelEvent) != nil {
		t.Error("Get should return nil for unopened channel")
	}

	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("test"))
	err := mux.Route(pkt)
	if err != nil {
		t.Fatalf("Route: %v", err)
	}

	stats := mux.Stats()
	if len(stats) != 1 {
		t.Errorf("Stats returned %d channels, want 1", len(stats))
	}

	mux.CloseAll()
	if mux.Get(ChannelRPC) != nil {
		t.Error("CloseAll should remove channels")
	}
}
