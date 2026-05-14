package server

import "errors"

var (
	// ErrServerNotInitialized is returned when server methods are called before initialization.
	ErrServerNotInitialized = errors.New("server not initialized, call Initialize first")
)
