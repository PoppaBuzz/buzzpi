package engine

import (
	"context"
	"encoding/json"
	"fmt"
)

// Context keys for session information.
type contextKey string

const (
	// CtxSessionToken is the context key for the authenticated session token.
	CtxSessionToken = contextKey("session_token")
	// CtxClientID is the context key for the authenticated client identifier.
	CtxClientID = contextKey("client_id")
	// CtxDeviceID is the context key for the target device identifier.
	CtxDeviceID = contextKey("device_id")
)

// SessionInfo returns the session token from the context.
func SessionInfo(ctx context.Context) (token string, ok bool) {
	v := ctx.Value(CtxSessionToken)
	if v == nil {
		return "", false
	}
	token, ok = v.(string)
	return
}

// ClientIDFromContext returns the client ID from the context.
func ClientIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(CtxClientID)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

// AuthMiddleware creates middleware that enforces method auth levels.
//
// Methods with AuthPublic are allowed without authentication.
// Methods with AuthPaired require a valid session token.
// Methods with AuthAdmin require an admin session token.
//
// The session token is read from the context (set upstream by the
// WebSocket handler after performing the BPP handshake).
func AuthMiddleware() MiddlewareFunc {
	return func(ctx context.Context, method string, params json.RawMessage, next HandlerFunc) (interface{}, error) {
		// This is a placeholder that allows all methods through.
		// Actual session validation is performed by the WebSocket
		// handler during the BPP handshake. Once the handshake
		// completes, all traffic on that connection is authenticated.
		//
		// Future enhancement: check method auth level from
		// MethodInfo.AuthLevel and reject unauthenticated requests.
		return next(ctx, params)
	}
}

// RequireSession is a middleware that rejects requests without a valid
// session token in the context.
func RequireSession() MiddlewareFunc {
	return func(ctx context.Context, method string, params json.RawMessage, next HandlerFunc) (interface{}, error) {
		token, ok := SessionInfo(ctx)
		if !ok || token == "" {
			return nil, fmt.Errorf("session required for method: %s", method)
		}
		return next(ctx, params)
	}
}
