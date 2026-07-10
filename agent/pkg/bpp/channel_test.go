package bpp

import (
	"io"
	"sync"
	"testing"
)

func TestNewChannel(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }

	t.Run("control channel defaults", func(t *testing.T) {
		ch := NewChannel(ChannelControl, ModeReliableOrdered, sendFn)
		if ch.ID() != ChannelControl {
			t.Errorf("ID = %d", ch.ID())
		}
		if ch.Mode() != ModeReliableOrdered {
			t.Errorf("Mode = %v", ch.Mode())
		}
	})

	t.Run("unreliable channel", func(t *testing.T) {
		ch := NewChannel(ChannelUnreliable, ModeUnreliableUnordered, sendFn)
		if ch.Mode() != ModeUnreliableUnordered {
			t.Errorf("Mode = %v", ch.Mode())
		}
	})

	t.Run("mode override via argument", func(t *testing.T) {
		// Default for ChannelRPC is ModeReliableOrdered; override to unreliable.
		ch := NewChannel(ChannelRPC, ModeUnreliableUnordered, sendFn)
		if ch.Mode() != ModeUnreliableUnordered {
			t.Errorf("Mode = %v, want unreliable", ch.Mode())
		}
	})
}

func TestDefaultChannelConfig(t *testing.T) {
	t.Parallel()

	t.Run("control is reliable ordered", func(t *testing.T) {
		cfg := DefaultChannelConfig(ChannelControl)
		if cfg.Mode != ModeReliableOrdered {
			t.Errorf("Mode = %v", cfg.Mode)
		}
		if cfg.ChannelID != ChannelControl {
			t.Errorf("ChannelID = %d", cfg.ChannelID)
		}
	})

	t.Run("rpc is reliable ordered", func(t *testing.T) {
		cfg := DefaultChannelConfig(ChannelRPC)
		if cfg.Mode != ModeReliableOrdered {
			t.Errorf("Mode = %v", cfg.Mode)
		}
	})

	t.Run("stream is stream mode", func(t *testing.T) {
		cfg := DefaultChannelConfig(ChannelStream)
		if cfg.Mode != ModeStream {
			t.Errorf("Mode = %v", cfg.Mode)
		}
	})

	t.Run("unreliable is unreliable", func(t *testing.T) {
		cfg := DefaultChannelConfig(ChannelUnreliable)
		if cfg.Mode != ModeUnreliableUnordered {
			t.Errorf("Mode = %v", cfg.Mode)
		}
	})
}

func TestChannelSendUnreliable(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var sent []*Packet
	sendFn := func(pkt *Packet) error {
		mu.Lock()
		sent = append(sent, pkt)
		mu.Unlock()
		return nil
	}

	ch := NewChannel(ChannelUnreliable, ModeUnreliableUnordered, sendFn)
	err := ch.Send([]byte("hello unreliable"))
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	mu.Lock()
	if len(sent) != 1 {
		t.Fatalf("sent %d packets, want 1", len(sent))
	}
	if string(sent[0].Payload) != "hello unreliable" {
		t.Errorf("payload = %q", string(sent[0].Payload))
	}
	mu.Unlock()
}

func TestChannelSendReliable(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var sent []*Packet
	sendFn := func(pkt *Packet) error {
		mu.Lock()
		sent = append(sent, pkt)
		mu.Unlock()
		return nil
	}

	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)
	err := ch.Send([]byte("reliable msg"))
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	mu.Lock()
	if len(sent) != 1 {
		t.Fatalf("sent %d packets, want 1", len(sent))
	}
	// Verify it was stored in pending for retransmit.
	if len(ch.pending) != 1 {
		t.Errorf("pending = %d, want 1", len(ch.pending))
	}
	pkt := sent[0]
	if pkt.Header.Sequence != 0 {
		t.Errorf("Sequence = %d, want 0", pkt.Header.Sequence)
	}
	mu.Unlock()
}

func TestChannelReceiveReliableOrdered(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var sentAcks []*Packet
	sendFn := func(pkt *Packet) error {
		mu.Lock()
		sentAcks = append(sentAcks, pkt)
		mu.Unlock()
		return nil
	}

	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)
	stats := ch.Stats()
	if stats.PacketsRecv != 0 {
		t.Error("initial PacketsRecv should be 0")
	}

	// Receive in-order.
	pkt1 := NewPacket(PacketTypeRequest, ChannelRPC, []byte("first"))
	pkt1.Header.Sequence = 0
	err := ch.Receive(pkt1)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}

	stats = ch.Stats()
	if stats.PacketsRecv != 1 {
		t.Errorf("PacketsRecv = %d, want 1", stats.PacketsRecv)
	}

	mu.Lock()
	if len(sentAcks) != 1 {
		t.Fatalf("sent %d acks, want 1", len(sentAcks))
	}
	if sentAcks[0].Header.PType != PacketTypeAck {
		t.Errorf("expected Ack packet, got %v", sentAcks[0].Header.PType)
	}
	mu.Unlock()
}

func TestChannelReceiveOutOfOrder(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)

	// Receive out-of-order (seq 1 before seq 0).
	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("second"))
	pkt.Header.Sequence = 1
	err := ch.Receive(pkt)
	if err == nil {
		t.Fatal("expected error for out-of-order packet")
	}

	stats := ch.Stats()
	if stats.PacketsLost != 1 {
		t.Errorf("PacketsLost = %d, want 1", stats.PacketsLost)
	}
}

func TestChannelReceiveReliableUnordered(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var ackCount int
	sendFn := func(pkt *Packet) error {
		mu.Lock()
		ackCount++
		mu.Unlock()
		return nil
	}

	ch := NewChannel(ChannelReliable, ModeReliableUnordered, sendFn)

	pkt1 := NewPacket(PacketTypeRequest, ChannelReliable, []byte("msg1"))
	pkt1.Header.Sequence = 5
	err := ch.Receive(pkt1)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}

	pkt2 := NewPacket(PacketTypeRequest, ChannelReliable, []byte("msg2"))
	pkt2.Header.Sequence = 2 // out of order — still delivers and acks
	err = ch.Receive(pkt2)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}

	mu.Lock()
	if ackCount != 2 {
		t.Errorf("acks = %d, want 2", ackCount)
	}
	mu.Unlock()
}

func TestChannelReceiveUnreliable(t *testing.T) {
	t.Parallel()

	ch := NewChannel(ChannelUnreliable, ModeUnreliableUnordered, nil)
	err := ch.Receive(NewPacket(PacketTypeRequest, ChannelUnreliable, []byte("data")))
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	// Unreliable receive is a no-op — should not error.
}

func TestChannelStreamRead(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelStream, ModeStream, sendFn)

	// Send data into the stream.
	go func() {
		pkt := NewPacket(PacketTypeRequest, ChannelStream, []byte("stream data"))
		err := ch.Receive(pkt)
		if err != nil {
			t.Errorf("Receive: %v", err)
		}
	}()

	// Read it back.
	buf := make([]byte, 64)
	n, err := ch.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "stream data" {
		t.Errorf("read = %q", string(buf[:n]))
	}

	stats := ch.Stats()
	if stats.StreamBytesRead != uint64(len("stream data")) {
		t.Errorf("StreamBytesRead = %d", stats.StreamBytesRead)
	}
}

func TestChannelStreamEOF(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelStream, ModeStream, sendFn)

	err := ch.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}

	buf := make([]byte, 64)
	n, err := ch.Read(buf)
	if err != io.EOF {
		t.Fatalf("expected EOF, got err=%v n=%d", err, n)
	}
	if n != 0 {
		t.Errorf("n = %d, want 0", n)
	}
}

func TestChannelReadOnNonStream(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelRPC, ModeReliableOrdered, sendFn)

	buf := make([]byte, 64)
	_, err := ch.Read(buf)
	if err == nil {
		t.Fatal("expected error for non-stream channel")
	}
}

func TestChannelCloseSafe(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelStream, ModeStream, sendFn)

	// Close multiple times should not panic.
	ch.Close()
	ch.Close()
}

func TestChannelStats(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelUnreliable, ModeUnreliableUnordered, sendFn)

	// Send a packet.
	_ = ch.Send([]byte("stats test"))
	stats := ch.Stats()
	if stats.PacketsSent != 1 {
		t.Errorf("PacketsSent = %d, want 1", stats.PacketsSent)
	}
	if stats.BytesSent < 8 {
		t.Errorf("BytesSent = %d, want >= 8", stats.BytesSent)
	}
}

func TestMultiplexer(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	mux := NewMultiplexer(sendFn)

	t.Run("open and get", func(t *testing.T) {
		ch := mux.Open(ChannelRPC, ModeReliableOrdered)
		if ch == nil {
			t.Fatal("Open returned nil")
		}
		if ch.ID() != ChannelRPC {
			t.Errorf("ID = %d", ch.ID())
		}

		// Get should return the same channel.
		ch2 := mux.Get(ChannelRPC)
		if ch2 != ch {
			t.Error("Get returned different channel")
		}
	})

	t.Run("open returns existing channel", func(t *testing.T) {
		ch1 := mux.Open(ChannelEvent, ModeReliableOrdered)
		ch2 := mux.Open(ChannelEvent, ModeUnreliableUnordered) // mode ignored
		if ch1 != ch2 {
			t.Error("Open should return existing channel")
		}
	})

	t.Run("get returns nil for unopened", func(t *testing.T) {
		ch := mux.Get(ChannelID(99))
		if ch != nil {
			t.Errorf("expected nil, got %v", ch)
		}
	})
}

func TestMultiplexerRoute(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	mux := NewMultiplexer(sendFn)

	// Open RPC channel.
	rpcCh := mux.Open(ChannelRPC, ModeReliableOrdered)

	// Route a packet to RPC.
	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("route test"))
	pkt.Header.Sequence = 0
	err := mux.Route(pkt)
	if err != nil {
		t.Fatalf("Route: %v", err)
	}

	_ = rpcCh // rpcCh received the packet
}

func TestMultiplexerAutoOpen(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	mux := NewMultiplexer(sendFn)

	// Route to control channel (≤ ChannelUnreliable) — should auto-open.
	pkt := NewPacket(PacketTypeHeartbeat, ChannelControl, nil)
	err := mux.Route(pkt)
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	ch := mux.Get(ChannelControl)
	if ch == nil {
		t.Fatal("ChannelControl should have been auto-opened")
	}
}

func TestMultiplexerRouteUnknownChannel(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	mux := NewMultiplexer(sendFn)

	pkt := NewPacket(PacketTypeRequest, ChannelID(99), []byte("data"))
	err := mux.Route(pkt)
	if err == nil {
		t.Fatal("expected error for unknown channel > ChannelUnreliable")
	}
}

func TestMultiplexerCloseAll(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	mux := NewMultiplexer(sendFn)

	mux.Open(ChannelRPC, ModeReliableOrdered)
	mux.Open(ChannelEvent, ModeReliableOrdered)

	mux.CloseAll()

	if ch := mux.Get(ChannelRPC); ch != nil {
		t.Error("channel should be removed after CloseAll")
	}
}

func TestMultiplexerStats(t *testing.T) {
	t.Parallel()

	sendFn := func(pkt *Packet) error { return nil }
	mux := NewMultiplexer(sendFn)

	mux.Open(ChannelRPC, ModeReliableOrdered)
	mux.Open(ChannelEvent, ModeReliableOrdered)

	stats := mux.Stats()
	if len(stats) != 2 {
		t.Errorf("stats len = %d, want 2", len(stats))
	}
}

func TestChannelSendConcurrent(t *testing.T) {
	// Test that concurrent Send calls do not cause data races.
	var mu sync.Mutex
	var sentCount int
	sendFn := func(pkt *Packet) error {
		mu.Lock()
		sentCount++
		mu.Unlock()
		return nil
	}

	ch := NewChannel(ChannelReliable, ModeReliableUnordered, sendFn)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ch.Send([]byte("concurrent"))
		}()
	}
	wg.Wait()

	mu.Lock()
	if sentCount != 10 {
		t.Errorf("sent %d packets, want 10", sentCount)
	}
	mu.Unlock()
}

func TestChannelModeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		m    ChannelMode
		want string
	}{
		{ModeReliableOrdered, "reliable_ordered"},
		{ModeReliableUnordered, "reliable_unordered"},
		{ModeUnreliableUnordered, "unreliable_unordered"},
		{ModeStream, "stream"},
		{ChannelMode(99), "unknown"},
	}
	for _, tt := range tests {
		got := tt.m.String()
		if got != tt.want {
			t.Errorf("ChannelMode(%d).String() = %q, want %q", tt.m, got, tt.want)
		}
	}
}

func TestChannelStreamDataRace(t *testing.T) {
	// Ensure stream reads and writes don't race.
	sendFn := func(pkt *Packet) error { return nil }
	ch := NewChannel(ChannelStream, ModeStream, sendFn)

	var wg sync.WaitGroup

	// Writer goroutines.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pkt := NewPacket(PacketTypeRequest, ChannelStream, []byte("hello"))
			_ = ch.Receive(pkt)
		}()
	}

	// Wait for all writers.
	wg.Wait()

	// Close the stream to unblock readers.
	ch.Close()

	// Reader goroutines - all should get data or EOF.
	var readWg sync.WaitGroup
	for i := 0; i < 3; i++ {
		readWg.Add(1)
		go func() {
			defer readWg.Done()
			buf := make([]byte, 64)
			for {
				_, err := ch.Read(buf)
				if err == io.EOF {
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}
	readWg.Wait()
}
