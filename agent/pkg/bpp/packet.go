// Package bpp implements the BuzzPi Protocol (BPP) wire format.
//
// BPP uses a binary framing layer over WebSocket for efficiency on constrained
// devices. Each message is prefixed with a compact fixed-size header followed
// by an optional variable-length payload.
//
// Wire format (big-endian):
//
//	┌──────────────────────────────────────────────┐
//	│ BPP Binary Frame                             │
//	├──────────┬───────────────────────────────────┤
//	│ Header   │ 16 bytes (fixed)                  │
//	│          │ - magic:    uint16 (0xB5 0x50)    │
//	│          │ - version:  uint8                 │
//	│          │ - flags:    uint8                 │
//	│          │ - ptype:    uint8                 │
//	│          │ - channel:  uint8                 │
//	│          │ - sequence: uint32                │
//	│          │ - length:   uint32                │
//	│          │ - checksum: uint16                │
//	├──────────┼───────────────────────────────────┤
//	│ Payload  │ length bytes (0 for control frames)│
//	└──────────┴───────────────────────────────────┘
//
// The JSON envelope layer (Envelope) is serialized into the binary payload
// and transported inside these frames.
package bpp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

// Protocol constants.
const (
	Magic         = 0xB550 // "B" "P" in hex — BP for BuzzPi Protocol
	CurrentVer    = 1
	HeaderLen     = 16      // fixed header size in bytes
	MaxPayloadLen = 1 << 24 // 16 MB max payload
)

// Magic bytes on the wire.
var MagicBytes = [2]byte{0xB5, 0x50}

// PacketType identifies the kind of BPP packet.
type PacketType uint8

const (
	// PacketTypeRequest is a method invocation expecting a response.
	PacketTypeRequest PacketType = iota
	// PacketTypeResponse is a successful reply to a Request.
	PacketTypeResponse
	// PacketTypeError is an error reply to a Request.
	PacketTypeError
	// PacketTypeEvent is a unilateral notification (no response expected).
	PacketTypeEvent
	// PacketTypeHeartbeat is a keep-alive ping.
	PacketTypeHeartbeat
	// PacketTypeHeartbeatAck is a keep-alive pong.
	PacketTypeHeartbeatAck
	// PacketTypeAck is a delivery confirmation for reliable channels.
	PacketTypeAck
	// PacketTypeClose is a graceful connection close signal.
	PacketTypeClose
)

// String returns the human-readable packet type name.
func (pt PacketType) String() string {
	switch pt {
	case PacketTypeRequest:
		return "request"
	case PacketTypeResponse:
		return "response"
	case PacketTypeError:
		return "error"
	case PacketTypeEvent:
		return "event"
	case PacketTypeHeartbeat:
		return "heartbeat"
	case PacketTypeHeartbeatAck:
		return "heartbeat_ack"
	case PacketTypeAck:
		return "ack"
	case PacketTypeClose:
		return "close"
	default:
		return fmt.Sprintf("unknown(%d)", pt)
	}
}

// Flag bits in the flags byte.
const (
	FlagCompressed   uint8 = 1 << 0 // payload is zstd-compressed
	FlagEncrypted    uint8 = 1 << 1 // payload is encrypted
	FlagFragment     uint8 = 1 << 2 // packet is a fragment (more follows)
	FlagLastFragment uint8 = 1 << 3 // last fragment of a fragmented message
	FlagPriority     uint8 = 1 << 4 // high-priority packet
)

// ChannelID represents a multiplexed data channel within a BPP connection.
type ChannelID uint8

const (
	ChannelControl    ChannelID = 0 // control messages (heartbeat, ack, close)
	ChannelRPC        ChannelID = 1 // request/response RPC calls
	ChannelEvent      ChannelID = 2 // event notifications
	ChannelStream     ChannelID = 3 // reliable stream data (files, terminal)
	ChannelReliable   ChannelID = 4 // reliable unordered data
	ChannelUnreliable ChannelID = 5 // unreliable unordered data (screen, audio)
)

// PacketHeader is the fixed-size BPP frame header.
type PacketHeader struct {
	Magic    uint16     // magic number (0xB550)
	Version  uint8      // protocol version
	Flags    uint8      // control flags (compression, encryption, fragment)
	PType    PacketType // packet type
	Channel  ChannelID  // multiplexed channel
	Sequence uint32     // sequence number (for ordering/reliability)
	Length   uint32     // payload length (big-endian)
	Checksum uint16     // CRC-16 of payload (0 if no payload)
}

// Packet is a complete BPP wire frame.
type Packet struct {
	Header  PacketHeader
	Payload []byte
}

// NewPacket creates a packet with the given type, channel, and payload.
func NewPacket(ptype PacketType, channel ChannelID, payload []byte) *Packet {
	if payload == nil {
		payload = []byte{}
	}
	h := PacketHeader{
		Magic:   Magic,
		Version: CurrentVer,
		PType:   ptype,
		Channel: channel,
		Length:  uint32(len(payload)),
	}
	if h.Length > 0 {
		h.Checksum = computeChecksum(payload)
	}
	return &Packet{Header: h, Payload: payload}
}

// NewControlPacket creates a packet with no payload (control frames).
func NewControlPacket(ptype PacketType) *Packet {
	return NewPacket(ptype, ChannelControl, nil)
}

// IsControl returns true if this packet has no payload.
func (p *Packet) IsControl() bool {
	return p.Header.Length == 0
}

// IsCompressed returns true if the payload is zstd-compressed.
func (p *Packet) IsCompressed() bool {
	return p.Header.Flags&FlagCompressed != 0
}

// IsEncrypted returns true if the payload is encrypted.
func (p *Packet) IsEncrypted() bool {
	return p.Header.Flags&FlagEncrypted != 0
}

// IsFragment returns true if this is a fragmented message (more fragments follow).
func (p *Packet) IsFragment() bool {
	return p.Header.Flags&FlagFragment != 0
}

// IsLastFragment returns true if this is the last fragment.
func (p *Packet) IsLastFragment() bool {
	return p.Header.Flags&FlagLastFragment != 0
}

// SetFlag sets a flag bit on the packet.
func (p *Packet) SetFlag(flag uint8) {
	p.Header.Flags |= flag
}

// Marshal serializes the packet into its binary wire format.
// Returns the complete byte slice ready for transmission.
func (p *Packet) Marshal() ([]byte, error) {
	if p.Header.Length != uint32(len(p.Payload)) {
		p.Header.Length = uint32(len(p.Payload))
		if p.Header.Length > 0 {
			p.Header.Checksum = computeChecksum(p.Payload)
		}
	}

	buf := make([]byte, HeaderLen+len(p.Payload))
	binary.BigEndian.PutUint16(buf[0:2], p.Header.Magic)
	buf[2] = p.Header.Version
	buf[3] = p.Header.Flags
	buf[4] = byte(p.Header.PType)
	buf[5] = byte(p.Header.Channel)
	binary.BigEndian.PutUint32(buf[6:10], p.Header.Sequence)
	binary.BigEndian.PutUint32(buf[10:14], p.Header.Length)
	binary.BigEndian.PutUint16(buf[14:16], p.Header.Checksum)
	copy(buf[16:], p.Payload)
	return buf, nil
}

// UnmarshalPacket deserializes a binary BPP packet from a byte slice.
// The slice must contain at least HeaderLen bytes.
func UnmarshalPacket(data []byte) (*Packet, error) {
	if len(data) < HeaderLen {
		return nil, fmt.Errorf("bpp: packet too short: %d < %d", len(data), HeaderLen)
	}

	h := PacketHeader{
		Magic:    binary.BigEndian.Uint16(data[0:2]),
		Version:  data[2],
		Flags:    data[3],
		PType:    PacketType(data[4]),
		Channel:  ChannelID(data[5]),
		Sequence: binary.BigEndian.Uint32(data[6:10]),
		Length:   binary.BigEndian.Uint32(data[10:14]),
		Checksum: binary.BigEndian.Uint16(data[14:16]),
	}

	if h.Magic != Magic {
		return nil, fmt.Errorf("bpp: invalid magic: 0x%04X", h.Magic)
	}
	if h.Version != CurrentVer {
		return nil, fmt.Errorf("bpp: unsupported version: %d", h.Version)
	}
	if h.Length > MaxPayloadLen {
		return nil, fmt.Errorf("bpp: payload too large: %d > %d", h.Length, MaxPayloadLen)
	}

	payloadLen := int(h.Length)
	if len(data) < HeaderLen+payloadLen {
		return nil, fmt.Errorf("bpp: truncated payload: need %d, have %d",
			HeaderLen+payloadLen, len(data))
	}

	payload := make([]byte, payloadLen)
	copy(payload, data[HeaderLen:HeaderLen+payloadLen])

	if h.Length > 0 {
		cs := computeChecksum(payload)
		if cs != h.Checksum {
			return nil, fmt.Errorf("bpp: checksum mismatch: 0x%04X != 0x%04X", cs, h.Checksum)
		}
	}

	return &Packet{Header: h, Payload: payload}, nil
}

// ReadPacket reads a complete BPP packet from an io.Reader.
func ReadPacket(r io.Reader) (*Packet, error) {
	headerBuf := make([]byte, HeaderLen)
	if _, err := io.ReadFull(r, headerBuf); err != nil {
		return nil, fmt.Errorf("bpp: read header: %w", err)
	}

	h := PacketHeader{
		Magic:    binary.BigEndian.Uint16(headerBuf[0:2]),
		Version:  headerBuf[2],
		Flags:    headerBuf[3],
		PType:    PacketType(headerBuf[4]),
		Channel:  ChannelID(headerBuf[5]),
		Sequence: binary.BigEndian.Uint32(headerBuf[6:10]),
		Length:   binary.BigEndian.Uint32(headerBuf[10:14]),
		Checksum: binary.BigEndian.Uint16(headerBuf[14:16]),
	}

	if h.Magic != Magic {
		return nil, fmt.Errorf("bpp: invalid magic: 0x%04X", h.Magic)
	}
	if h.Version != CurrentVer {
		return nil, fmt.Errorf("bpp: unsupported version: %d", h.Version)
	}
	if h.Length > MaxPayloadLen {
		return nil, fmt.Errorf("bpp: payload too large: %d > %d", h.Length, MaxPayloadLen)
	}

	payload := make([]byte, h.Length)
	if h.Length > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, fmt.Errorf("bpp: read payload: %w", err)
		}
	}

	if h.Length > 0 {
		cs := computeChecksum(payload)
		if cs != h.Checksum {
			return nil, fmt.Errorf("bpp: checksum mismatch: 0x%04X != 0x%04X", cs, h.Checksum)
		}
	}

	return &Packet{Header: h, Payload: payload}, nil
}

// WriteTo writes the marshalled packet to an io.Writer.
func (p *Packet) WriteTo(w io.Writer) (int64, error) {
	data, err := p.Marshal()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}

// Fragment splits a large payload into multiple fragment packets.
func Fragment(ptype PacketType, channel ChannelID, payload []byte, maxSize int) ([]*Packet, error) {
	if maxSize < HeaderLen+1 {
		return nil, errors.New("bpp: max fragment size too small")
	}
	maxPayload := maxSize - HeaderLen
	if maxPayload <= 0 {
		return nil, errors.New("bpp: max fragment size must exceed header size")
	}

	if len(payload) <= maxPayload {
		// Single packet, no fragmentation needed.
		return []*Packet{NewPacket(ptype, channel, payload)}, nil
	}

	var packets []*Packet
	offset := 0
	for offset < len(payload) {
		end := offset + maxPayload
		if end > len(payload) {
			end = len(payload)
		}
		chunk := payload[offset:end]
		pkt := NewPacket(ptype, channel, chunk)
		pkt.SetFlag(FlagFragment)
		packets = append(packets, pkt)
		offset = end
	}
	// Mark the last fragment.
	if len(packets) > 0 {
		packets[len(packets)-1].SetFlag(FlagLastFragment)
	}
	return packets, nil
}

// Assemble reassembles fragment packets into a single payload.
func Assemble(fragments []*Packet) ([]byte, error) {
	if len(fragments) == 0 {
		return nil, errors.New("bpp: no fragments to assemble")
	}
	totalLen := 0
	for _, f := range fragments {
		totalLen += len(f.Payload)
	}

	result := make([]byte, 0, totalLen)
	for _, f := range fragments {
		result = append(result, f.Payload...)
	}
	return result, nil
}

// computeChecksum returns a CRC-16 (CCITT) of the data.
func computeChecksum(data []byte) uint16 {
	// Use CRC-32 sliced-by-4, drop to lower 16 bits.
	// This is fast and good enough for frame-level corruption detection.
	// CRC-16 table is ~64KB; CRC-32 is available in the stdlib.
	t := crc32.MakeTable(crc32.Castagnoli)
	return uint16(crc32.Checksum(data, t))
}

// String returns a human-readable representation of the packet header.
func (h PacketHeader) String() string {
	return fmt.Sprintf("magic=0x%04X ver=%d flags=0x%02X type=%s channel=%d len=%d csum=0x%04X",
		h.Magic, h.Version, h.Flags, h.PType, h.Channel, h.Length, h.Checksum)
}

// String returns a human-readable representation of the packet.
func (p *Packet) String() string {
	return fmt.Sprintf("Packet{%s payload=%d bytes}", p.Header.String(), len(p.Payload))
}

// Ensure Packet implements io.WriterTo.
var _ io.WriterTo = (*Packet)(nil)
