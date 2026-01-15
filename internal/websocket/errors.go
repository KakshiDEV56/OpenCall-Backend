package websocket

import "errors"

var (
	ErrMessageBufferFull = errors.New("message buffer is full")
	ErrSessionNotFound   = errors.New("session not found")
	ErrClientNotFound    = errors.New("client not found")
)
