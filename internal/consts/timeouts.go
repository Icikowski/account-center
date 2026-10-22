package consts

import "time"

// HTTP client & server timeouts and limits.
const (
	ServerTimeoutReadHeader = 5 * time.Second
	ServerTimeoutRead       = 15 * time.Second
	ServerTimeoutWrite      = 30 * time.Second
	ServerTimeoutIdle       = 60 * time.Second

	ServerMaxHeaderBytes = 64 << 10

	ClientTimeoutOIDC = 15 * time.Second
)
