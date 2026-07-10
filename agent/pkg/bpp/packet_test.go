package bpp

import (
	"bytes"
	"testing"
)

func TestNewPacket(t *testing.T) {
	t.Parallel()

	t.Run("with payload", func(t *testing.T) {
		pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("hello"))
		if pkt.Header.Magic != Magic {
			t.Errorf("Magic = 0x%04X, want 0x%04X", pkt.Header.Magic, Magic)
		}
		if pkt.Header.Version != CurrentVer {
			t.Errorf("Version = %d, want %d", pkt.Header.Version, CurrentVer)
		}
		if pkt.Header.PType != PacketTypeRequest {
			t.Errorf("PType = %v, want %v", pkt.Header.PType, PacketTypeRequest)
		}
		if pkt.Header.Channel != ChannelRPC {
			t.Errorf("Channel = %d, want %d", pkt.Header.Channel, ChannelRPC)
		}
		if pkt.Header.Length != 5 {
			t.Errorf("Length = %d, want 5", pkt.Header.Length)
		}
		if pkt.Header.Checksum == 0 {
			t.Error("Checksum should not be zero for payload")
		}
	})

	t.Run("nil payload becomes empty", func(t *testing.T) {
		pkt := NewPacket(PacketTypeHeartbeat, ChannelControl, nil)
		if pkt.Header.Length != 0 {
			t.Errorf("Length = %d, want 0", pkt.Header.Length)
		}
		if pkt.Header.Checksum != 0 {
			t.Errorf("Checksum = %d, want 0 for empty payload", pkt.Header.Checksum)
		}
	})

	t.Run("control packet has no payload", func(t *testing.T) {
		pkt := NewControlPacket(PacketTypeClose)
		if !pkt.IsControl() {
			t.Error("expected control packet")
		}
	})
}

func TestPacketFlags(t *testing.T) {
	t.Parallel()

	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("data"))
	if pkt.IsCompressed() {
		t.Error("not compressed")
	}
	if pkt.IsEncrypted() {
		t.Error("not encrypted")
	}
	if pkt.IsFragment() {
		t.Error("not a fragment")
	}
	if pkt.IsLastFragment() {
		t.Error("not last fragment")
	}

	pkt.SetFlag(FlagCompressed)
	if !pkt.IsCompressed() {
		t.Error("should be compressed")
	}

	pkt.SetFlag(FlagFragment)
	if !pkt.IsFragment() {
		t.Error("should be fragment")
	}

	pkt.SetFlag(FlagLastFragment)
	if !pkt.IsLastFragment() {
		t.Error("should be last fragment")
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("small payload", func(t *testing.T) {
		payload := []byte("hello bpp")
		original := NewPacket(PacketTypeRequest, ChannelRPC, payload)
		original.Header.Sequence = 42

		data, err := original.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if len(data) != HeaderLen+len(payload) {
			t.Fatalf("wire size = %d, want %d", len(data), HeaderLen+len(payload))
		}

		parsed, err := UnmarshalPacket(data)
		if err != nil {
			t.Fatalf("UnmarshalPacket: %v", err)
		}
		if parsed.Header.Magic != Magic {
			t.Errorf("Magic = 0x%04X", parsed.Header.Magic)
		}
		if parsed.Header.Version != CurrentVer {
			t.Errorf("Version = %d", parsed.Header.Version)
		}
		if parsed.Header.PType != PacketTypeRequest {
			t.Errorf("PType = %v", parsed.Header.PType)
		}
		if parsed.Header.Channel != ChannelRPC {
			t.Errorf("Channel = %d", parsed.Header.Channel)
		}
		if parsed.Header.Sequence != 42 {
			t.Errorf("Sequence = %d, want 42", parsed.Header.Sequence)
		}
		if parsed.Header.Length != 9 {
			t.Errorf("Length = %d, want 9", parsed.Header.Length)
		}
		if string(parsed.Payload) != "hello bpp" {
			t.Errorf("Payload = %q, want %q", string(parsed.Payload), "hello bpp")
		}
	})

	t.Run("empty payload", func(t *testing.T) {
		pkt := NewControlPacket(PacketTypeHeartbeat)
		data, err := pkt.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		parsed, err := UnmarshalPacket(data)
		if err != nil {
			t.Fatalf("UnmarshalPacket: %v", err)
		}
		if !parsed.IsControl() {
			t.Error("expected control packet")
		}
		if parsed.Header.PType != PacketTypeHeartbeat {
			t.Errorf("PType = %v", parsed.Header.PType)
		}
	})
}

func TestMarshalSynchronizesLength(t *testing.T) {
	t.Parallel()

	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("original"))
	pkt.Header.Length = 999 // wrong length

	data, err := pkt.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	parsed, err := UnmarshalPacket(data)
	if err != nil {
		t.Fatalf("UnmarshalPacket: %v", err)
	}
	if parsed.Header.Length != 8 {
		t.Errorf("Length corrected to %d, want 8", parsed.Header.Length)
	}
}

func TestUnmarshalPacketErrors(t *testing.T) {
	t.Parallel()

	t.Run("too short", func(t *testing.T) {
		_, err := UnmarshalPacket([]byte{0, 1, 2})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid magic", func(t *testing.T) {
		buf := make([]byte, HeaderLen)
		buf[0] = 0x00
		buf[1] = 0x01
		_, err := UnmarshalPacket(buf)
		if err == nil {
			t.Fatal("expected error for invalid magic")
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		buf := make([]byte, HeaderLen)
		buf[0] = 0xB5
		buf[1] = 0x50
		buf[2] = 99 // version
		_, err := UnmarshalPacket(buf)
		if err == nil {
			t.Fatal("expected error for unsupported version")
		}
	})

	t.Run("payload too large", func(t *testing.T) {
		buf := make([]byte, HeaderLen)
		buf[0] = 0xB5
		buf[1] = 0x50
		buf[2] = CurrentVer
		binaryBuf := bytes.NewBuffer(buf[10:14])
		binaryBuf.Reset()
		// Set length past max
		// Actually we need to write big-endian uint32
		buf[10] = 0x02 // > MaxPayloadLen (16MB)
		buf[11] = 0x00
		buf[12] = 0x00
		buf[13] = 0x00
		_, err := UnmarshalPacket(buf)
		if err == nil {
			t.Fatal("expected error for too-large payload")
		}
	})

	t.Run("truncated payload", func(t *testing.T) {
		// Create a valid header claiming 10 bytes payload but only provide 5.
		buf := make([]byte, HeaderLen+5)
		buf[0] = 0xB5
		buf[1] = 0x50
		buf[2] = CurrentVer
		buf[10] = 0
		buf[11] = 0
		buf[12] = 0
		buf[13] = 10 // length = 10
		_, err := UnmarshalPacket(buf)
		if err == nil {
			t.Fatal("expected error for truncated payload")
		}
	})
}

func TestChecksum(t *testing.T) {
	t.Parallel()

	payload := []byte("checksum test data")
	pkt := NewPacket(PacketTypeRequest, ChannelRPC, payload)

	data, err := pkt.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Corrupt the checksum.
	data[15] ^= 0xFF
	_, err = UnmarshalPacket(data)
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
}

func TestFragmentAndAssemble(t *testing.T) {
	t.Parallel()

	t.Run("small payload no fragmentation", func(t *testing.T) {
		payload := []byte("small")
		packets, err := Fragment(PacketTypeRequest, ChannelStream, payload, HeaderLen+10)
		if err != nil {
			t.Fatalf("Fragment: %v", err)
		}
		if len(packets) != 1 {
			t.Fatalf("got %d packets, want 1", len(packets))
		}
		if packets[0].IsFragment() {
			t.Error("single packet should not be fragment")
		}
	})

	t.Run("large payload splits into fragments", func(t *testing.T) {
		payload := make([]byte, 100)
		for i := range payload {
			payload[i] = byte(i)
		}
		maxSize := HeaderLen + 30
		packets, err := Fragment(PacketTypeRequest, ChannelStream, payload, maxSize)
		if err != nil {
			t.Fatalf("Fragment: %v", err)
		}
		if len(packets) < 2 {
			t.Fatalf("got %d packets, expected at least 2", len(packets))
		}
		// All but last should have FlagFragment.
		for i, pkt := range packets[:len(packets)-1] {
			if !pkt.IsFragment() {
				t.Errorf("packet[%d] missing FlagFragment", i)
			}
			if pkt.IsLastFragment() {
				t.Errorf("packet[%d] should not be last fragment", i)
			}
		}
		// Last packet must have FlagLastFragment (and FlagFragment).
		last := packets[len(packets)-1]
		if !last.IsLastFragment() {
			t.Error("last packet missing FlagLastFragment")
		}
		if !last.IsFragment() {
			t.Error("last packet missing FlagFragment")
		}
	})

	t.Run("assemble reconstructs payload", func(t *testing.T) {
		payload := []byte("hello world this is a fragmented message")
		packets, err := Fragment(PacketTypeRequest, ChannelStream, payload, HeaderLen+10)
		if err != nil {
			t.Fatalf("Fragment: %v", err)
		}
		result, err := Assemble(packets)
		if err != nil {
			t.Fatalf("Assemble: %v", err)
		}
		if string(result) != string(payload) {
			t.Errorf("reconstructed = %q, want %q", string(result), string(payload))
		}
	})

	t.Run("assemble empty slice returns error", func(t *testing.T) {
		_, err := Assemble([]*Packet{})
		if err == nil {
			t.Fatal("expected error for empty fragments")
		}
	})
}

func TestFragmentEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("maxSize too small", func(t *testing.T) {
		_, err := Fragment(PacketTypeRequest, ChannelRPC, []byte("data"), HeaderLen)
		if err == nil {
			t.Fatal("expected error for maxSize <= HeaderLen")
		}
	})

	t.Run("exact fit", func(t *testing.T) {
		payload := make([]byte, 10)
		packets, err := Fragment(PacketTypeRequest, ChannelRPC, payload, HeaderLen+10)
		if err != nil {
			t.Fatalf("Fragment: %v", err)
		}
		if len(packets) != 1 {
			t.Errorf("got %d packets, want 1", len(packets))
		}
	})
}

func TestWriteTo(t *testing.T) {
	t.Parallel()

	pkt := NewPacket(PacketTypeResponse, ChannelRPC, []byte("response data"))
	var buf bytes.Buffer
	n, err := pkt.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if n != int64(HeaderLen+13) {
		t.Errorf("wrote %d bytes, want %d", n, HeaderLen+13)
	}
	parsed, err := UnmarshalPacket(buf.Bytes())
	if err != nil {
		t.Fatalf("UnmarshalPacket after WriteTo: %v", err)
	}
	if string(parsed.Payload) != "response data" {
		t.Errorf("Payload = %q", string(parsed.Payload))
	}
}

func TestReadPacket(t *testing.T) {
	t.Parallel()

	t.Run("round-trip", func(t *testing.T) {
		pkt := NewPacket(PacketTypeEvent, ChannelEvent, []byte("event data"))
		data, err := pkt.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		reader := bytes.NewReader(data)
		parsed, err := ReadPacket(reader)
		if err != nil {
			t.Fatalf("ReadPacket: %v", err)
		}
		if parsed.Header.PType != PacketTypeEvent {
			t.Errorf("PType = %v", parsed.Header.PType)
		}
		if string(parsed.Payload) != "event data" {
			t.Errorf("Payload = %q", string(parsed.Payload))
		}
	})

	t.Run("truncated header", func(t *testing.T) {
		reader := bytes.NewReader([]byte{0xB5, 0x50})
		_, err := ReadPacket(reader)
		if err == nil {
			t.Fatal("expected error for truncated header")
		}
	})
}

func TestPacketString(t *testing.T) {
	t.Parallel()

	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("test"))
	s := pkt.String()
	if s == "" {
		t.Fatal("String() returned empty")
	}
	hs := pkt.Header.String()
	if hs == "" {
		t.Fatal("Header.String() returned empty")
	}
}

func TestPacketTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pt   PacketType
		want string
	}{
		{PacketTypeRequest, "request"},
		{PacketTypeResponse, "response"},
		{PacketTypeError, "error"},
		{PacketTypeEvent, "event"},
		{PacketTypeHeartbeat, "heartbeat"},
		{PacketTypeHeartbeatAck, "heartbeat_ack"},
		{PacketTypeAck, "ack"},
		{PacketTypeClose, "close"},
		{PacketType(255), "unknown(255)"},
	}
	for _, tt := range tests {
		got := tt.pt.String()
		if got != tt.want {
			t.Errorf("PacketType(%d).String() = %q, want %q", tt.pt, got, tt.want)
		}
	}
}
