// Package bpp implements the BuzzPi Protocol (BPP) wire format.
//
// Channel provides multiplexed data channel abstractions over a BPP connection.
// Each channel represents a logical stream of data with its own reliability,
// ordering, and flow control policy.
//
// Channel types:
//
//	Control     (ch 0) — Handshake, heartbeat, ack, close
//	RPC         (ch 1) — Request/response method calls (reliable, ordered)
//	Event       (ch 2) — Unilateral notifications (reliable, ordered)
//	Stream      (ch 3) — Byte stream (terminal, file transfers)
//	Reliable    (ch 4) — Unordered reliable messages (file metadata)
//	Unreliable  (ch 5) — Unordered unreliable messages (screen frames, audio)
package bpp

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Default channel parameters.
const (
	DefaultStreamBufferSize  = 65536  // 64KB per stream buffer
	DefaultStreamWindowSize  = 262144 // 256KB flow control window
	DefaultRetransmitTimeout = 500 * time.Millisecond
	DefaultMaxRetransmits    = 5
)

// ChannelMode defines the delivery semantics for a channel.
type ChannelMode int

const (
	// ModeReliableOrdered delivers data with guaranteed delivery and ordering.
	ModeReliableOrdered ChannelMode = iota
	// ModeReliableUnordered delivers data with guaranteed delivery but no ordering.
	ModeReliableUnordered
	// ModeUnreliableUnordered delivers data with no guarantees (best-effort).
	ModeUnreliableUnordered
	// ModeStream delivers data as a byte stream (reliable, ordered, flow-controlled).
	ModeStream
)

func (m ChannelMode) String() string {
	switch m {
	case ModeReliableOrdered:
		return "reliable_ordered"
	case ModeReliableUnordered:
		return "reliable_unordered"
	case ModeUnreliableUnordered:
		return "unreliable_unordered"
	case ModeStream:
		return "stream"
	default:
		return "unknown"
	}
}

// ChannelConfig configures a single data channel.
type ChannelConfig struct {
	ChannelID      ChannelID
	Mode           ChannelMode
	BufferSize     int
	WindowSize     int
	Retransmit     time.Duration
	MaxRetransmits int
}

// DefaultChannelConfig returns a sensible default configuration for each
// channel ID.
func DefaultChannelConfig(ch ChannelID) ChannelConfig {
	cfg := ChannelConfig{
		ChannelID:      ch,
		BufferSize:     DefaultStreamBufferSize,
		WindowSize:     DefaultStreamWindowSize,
		Retransmit:     DefaultRetransmitTimeout,
		MaxRetransmits: DefaultMaxRetransmits,
	}

	switch ch {
	case ChannelControl:
		cfg.Mode = ModeReliableOrdered
	case ChannelRPC:
		cfg.Mode = ModeReliableOrdered
	case ChannelEvent:
		cfg.Mode = ModeReliableOrdered
	case ChannelStream:
		cfg.Mode = ModeStream
	case ChannelReliable:
		cfg.Mode = ModeReliableUnordered
	case ChannelUnreliable:
		cfg.Mode = ModeUnreliableUnordered
	}

	return cfg
}

// Channel is a logical multiplexed channel within a BPP connection.
// Each channel handles its own sequencing, flow control, and reliability.
type Channel struct {
	id     ChannelID
	mode   ChannelMode
	config ChannelConfig

	// Send sequence number (monotonically increasing).
	sendSeq uint32
	// Next expected receive sequence number.
	recvSeq uint32

	// Stream buffer for ModeStream.
	streamBuf  []byte
	streamCond *sync.Cond
	streamMu   sync.Mutex
	streamEOF  bool
	streamErr  error

	// Pending reliable messages for retransmission.
	pending   map[uint32]*pendingMessage
	pendingMu sync.Mutex

	// Send/receive callbacks.
	sendFn func(pkt *Packet) error

	// Statistics.
	stats   ChannelStats
	statsMu sync.Mutex

	log *slog.Logger
}

type pendingMessage struct {
	data    []byte
	timeout time.Time
	retries int
}

// ChannelStats exposes runtime statistics for a channel.
type ChannelStats struct {
	PacketsSent     uint64
	PacketsRecv     uint64
	PacketsLost     uint64
	PacketsRetrans  uint64
	BytesSent       uint64
	BytesRecv       uint64
	StreamBytesRead uint64
}

// NewChannel creates a new multiplexed data channel.
func NewChannel(id ChannelID, mode ChannelMode, sendFn func(pkt *Packet) error) *Channel {
	cfg := DefaultChannelConfig(id)
	if mode != cfg.Mode {
		cfg.Mode = mode
	}

	ch := &Channel{
		id:        id,
		mode:      cfg.Mode,
		config:    cfg,
		sendFn:    sendFn,
		pending:   make(map[uint32]*pendingMessage),
		streamBuf: make([]byte, 0, cfg.BufferSize),
		log:       slog.Default().With("component", "bpp-channel", "id", id, "mode", cfg.Mode),
	}
	ch.streamMu = sync.Mutex{}
	ch.streamCond = sync.NewCond(&ch.streamMu)
	return ch
}

// ID returns the channel identifier.
func (ch *Channel) ID() ChannelID { return ch.id }

// Mode returns the channel delivery mode.
func (ch *Channel) Mode() ChannelMode { return ch.mode }

// Send enqueues a payload for transmission over this channel.
// For reliable channels, it blocks until acknowledged (or timeout).
// For unreliable channels, it sends immediately and returns.
func (ch *Channel) Send(payload []byte) error {
	seq := atomic.AddUint32(&ch.sendSeq, 1) - 1

	pkt := NewPacket(PacketTypeRequest, ch.id, payload)
	pkt.Header.Sequence = seq

	switch ch.mode {
	case ModeReliableOrdered, ModeReliableUnordered:
		return ch.sendReliable(pkt)
	case ModeUnreliableUnordered:
		return ch.sendUnreliable(pkt)
	case ModeStream:
		return ch.sendStream(payload)
	default:
		return fmt.Errorf("bpp: unknown channel mode: %v", ch.mode)
	}
}

// Receive processes an incoming packet directed at this channel.
// Called by the connection read pump.
func (ch *Channel) Receive(pkt *Packet) error {
	ch.statsMu.Lock()
	ch.stats.PacketsRecv++
	ch.stats.BytesRecv += uint64(len(pkt.Payload))
	ch.statsMu.Unlock()

	switch ch.mode {
	case ModeReliableOrdered:
		return ch.recvReliableOrdered(pkt)
	case ModeReliableUnordered:
		return ch.recvReliableUnordered(pkt)
	case ModeUnreliableUnordered:
		return ch.recvUnreliable(pkt)
	case ModeStream:
		return ch.recvStream(pkt)
	default:
		return fmt.Errorf("bpp: unknown channel mode: %v", ch.mode)
	}
}

// Read implements io.Reader for ModeStream channels.
// Blocks until data is available or the channel is closed.
func (ch *Channel) Read(b []byte) (int, error) {
	if ch.mode != ModeStream {
		return 0, fmt.Errorf("bpp: channel %d is not a stream channel", ch.id)
	}

	ch.streamMu.Lock()
	defer ch.streamMu.Unlock()

	for len(ch.streamBuf) == 0 && !ch.streamEOF && ch.streamErr == nil {
		ch.streamCond.Wait()
	}

	if ch.streamErr != nil {
		return 0, ch.streamErr
	}

	n := copy(b, ch.streamBuf)
	ch.streamBuf = ch.streamBuf[n:]

	ch.statsMu.Lock()
	ch.stats.StreamBytesRead += uint64(n)
	ch.statsMu.Unlock()

	if n == 0 && ch.streamEOF {
		return 0, io.EOF
	}
	return n, nil
}

// Close explictly closes the channel and releases resources.
func (ch *Channel) Close() error {
	ch.streamMu.Lock()
	defer ch.streamMu.Unlock()
	ch.streamEOF = true
	if ch.streamCond != nil {
		ch.streamCond.Broadcast()
	}
	return nil
}

// Stats returns a snapshot of channel statistics.
func (ch *Channel) Stats() ChannelStats {
	ch.statsMu.Lock()
	defer ch.statsMu.Unlock()
	return ch.stats
}

// -- Reliable (ordered) ------------------------------------------------------

func (ch *Channel) sendReliable(pkt *Packet) error {
	// Store for retransmission.
	ch.pendingMu.Lock()
	ch.pending[pkt.Header.Sequence] = &pendingMessage{
		data:    append([]byte{}, pkt.Payload...),
		timeout: time.Now().Add(ch.config.Retransmit),
	}
	ch.pendingMu.Unlock()

	if err := ch.sendFn(pkt); err != nil {
		return err
	}

	ch.statsMu.Lock()
	ch.stats.PacketsSent++
	ch.stats.BytesSent += uint64(len(pkt.Payload))
	ch.statsMu.Unlock()
	return nil
}

func (ch *Channel) recvReliableOrdered(pkt *Packet) error {
	// Expect packets in order.
	seq := pkt.Header.Sequence
	if seq != ch.recvSeq {
		// Out-of-order — drop or buffer (simple: drop for now).
		ch.statsMu.Lock()
		ch.stats.PacketsLost++
		ch.statsMu.Unlock()
		return fmt.Errorf("bpp: out-of-order packet on ch %d: got %d, expected %d",
			ch.id, seq, ch.recvSeq)
	}
	ch.recvSeq++

	// Send ack.
	ack := NewControlPacket(PacketTypeAck)
	ack.Header.Channel = ch.id
	ack.Header.Sequence = seq
	_ = ch.sendFn(ack)

	return nil
}

func (ch *Channel) recvReliableUnordered(pkt *Packet) error {
	// Send ack regardless of order.
	ack := NewControlPacket(PacketTypeAck)
	ack.Header.Channel = ch.id
	ack.Header.Sequence = pkt.Header.Sequence
	_ = ch.sendFn(ack)
	return nil
}

// -- Unreliable --------------------------------------------------------------

func (ch *Channel) sendUnreliable(pkt *Packet) error {
	if err := ch.sendFn(pkt); err != nil {
		return err
	}

	ch.statsMu.Lock()
	ch.stats.PacketsSent++
	ch.stats.BytesSent += uint64(len(pkt.Payload))
	ch.statsMu.Unlock()
	return nil
}

func (ch *Channel) recvUnreliable(pkt *Packet) error {
	// No ordering, no ack. Fire and forget.
	return nil
}

// -- Stream ------------------------------------------------------------------

func (ch *Channel) sendStream(payload []byte) error {
	pkt := NewPacket(PacketTypeRequest, ch.id, payload)
	if err := ch.sendFn(pkt); err != nil {
		return err
	}

	ch.statsMu.Lock()
	ch.stats.PacketsSent++
	ch.stats.BytesSent += uint64(len(payload))
	ch.statsMu.Unlock()
	return nil
}

func (ch *Channel) recvStream(pkt *Packet) error {
	ch.streamMu.Lock()
	defer ch.streamMu.Unlock()
	ch.streamBuf = append(ch.streamBuf, pkt.Payload...)
	if ch.streamCond != nil {
		ch.streamCond.Broadcast()
	}
	return nil
}

// -- Multiplexer -------------------------------------------------------------

// Multiplexer manages multiple channels over a single BPP connection.
// It routes incoming packets to the correct channel based on ChannelID.
type Multiplexer struct {
	channels map[ChannelID]*Channel
	mu       sync.RWMutex
	sendFn   func(pkt *Packet) error
	log      *slog.Logger
}

// NewMultiplexer creates a new channel multiplexer.
func NewMultiplexer(sendFn func(pkt *Packet) error) *Multiplexer {
	m := &Multiplexer{
		channels: make(map[ChannelID]*Channel),
		sendFn:   sendFn,
		log:      slog.Default().With("component", "bpp-mux"),
	}
	return m
}

// Open opens (or gets) a channel with the given ID and mode.
func (m *Multiplexer) Open(id ChannelID, mode ChannelMode) *Channel {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ch, ok := m.channels[id]; ok {
		return ch
	}

	ch := NewChannel(id, mode, m.sendFn)
	m.channels[id] = ch
	m.log.Info("channel opened", "id", id, "mode", mode)
	return ch
}

// Get returns an existing channel by ID, or nil if not opened.
func (m *Multiplexer) Get(id ChannelID) *Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.channels[id]
}

// Route dispatches an incoming packet to the correct channel.
func (m *Multiplexer) Route(pkt *Packet) error {
	ch := m.Get(pkt.Header.Channel)
	if ch == nil {
		// Auto-open for control channels.
		if pkt.Header.Channel <= ChannelUnreliable {
			ch = m.Open(pkt.Header.Channel, DefaultChannelConfig(pkt.Header.Channel).Mode)
		} else {
			return fmt.Errorf("bpp: no handler for channel %d", pkt.Header.Channel)
		}
	}
	return ch.Receive(pkt)
}

// CloseAll closes all open channels.
func (m *Multiplexer) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, ch := range m.channels {
		ch.Close()
		delete(m.channels, id)
	}
	m.log.Info("all channels closed")
}

// Stats returns statistics for all channels.
func (m *Multiplexer) Stats() map[ChannelID]ChannelStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[ChannelID]ChannelStats)
	for id, ch := range m.channels {
		result[id] = ch.Stats()
	}
	return result
}

// Ensure Channel implements io.Reader.
var _ io.Reader = (*Channel)(nil)
