// Package bpp defines the BuzzPi Protocol (BPP) message types.
// BPP is a JSON-based protocol over WebSocket for device-client communication.
package bpp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Protocol version constants.
const (
	CurrentVersion = 1
)

// Message types.
const (
	TypeRequest  = "request"
	TypeResponse = "response"
	TypeEvent    = "event"
	TypeError    = "error"
)

// Envelope is the universal wrapper for all BPP messages.
type Envelope struct {
	Version int             `json:"v"`
	ID      string          `json:"id"`
	TS      string          `json:"ts"`
	Type    string          `json:"type"`
	Method  string          `json:"method,omitempty"`
	RID     string          `json:"rid,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ErrorObj       `json:"error,omitempty"`
}

// ErrorObj represents a BPP error.
type ErrorObj struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *ErrorObj) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewRequest creates a new request envelope.
func NewRequest(method string, params interface{}) (*Envelope, error) {
	var raw json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		raw = b
	}

	return &Envelope{
		Version: CurrentVersion,
		ID:      "msg_" + uuid.New().String()[:12],
		TS:      time.Now().UTC().Format(time.RFC3339),
		Type:    TypeRequest,
		Method:  method,
		RID:     "req_" + uuid.New().String()[:12],
		Params:  raw,
	}, nil
}

// NewResponse creates a response envelope for a given request.
func NewResponse(req *Envelope, result interface{}) (*Envelope, error) {
	var raw json.RawMessage
	if result != nil {
		b, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("marshal result: %w", err)
		}
		raw = b
	}

	return &Envelope{
		Version: CurrentVersion,
		ID:      "msg_" + uuid.New().String()[:12],
		TS:      time.Now().UTC().Format(time.RFC3339),
		Type:    TypeResponse,
		RID:     req.RID,
		Result:  raw,
	}, nil
}

// NewErrorResponse creates an error response envelope.
func NewErrorResponse(req *Envelope, code, message string) *Envelope {
	return &Envelope{
		Version: CurrentVersion,
		ID:      "msg_" + uuid.New().String()[:12],
		TS:      time.Now().UTC().Format(time.RFC3339),
		Type:    TypeError,
		RID:     req.RID,
		Error: &ErrorObj{
			Code:    code,
			Message: message,
		},
	}
}

// NewEvent creates an event envelope.
func NewEvent(method string, params interface{}) (*Envelope, error) {
	var raw json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal event params: %w", err)
		}
		raw = b
	}

	return &Envelope{
		Version: CurrentVersion,
		ID:      "msg_" + uuid.New().String()[:12],
		TS:      time.Now().UTC().Format(time.RFC3339),
		Type:    TypeEvent,
		Method:  method,
		Params:  raw,
	}, nil
}

// Marshal serializes the envelope to JSON.
func (e *Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal deserializes JSON into an envelope.
func Unmarshal(data []byte) (*Envelope, error) {
	var e Envelope
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}
	if e.Version != CurrentVersion {
		return nil, fmt.Errorf("unsupported protocol version: %d", e.Version)
	}
	return &e, nil
}
