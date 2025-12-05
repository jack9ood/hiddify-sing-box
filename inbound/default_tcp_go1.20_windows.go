//go:build windows && go1.20

package inbound

import (
	"context"
	"net"
)

const go120Available = true

func listenTFO(listenConfig net.ListenConfig, ctx context.Context, network string, address string) (net.Listener, error) {
	// TCP Fast Open is not supported on Windows due to compatibility issues
	// Fall back to regular Listen
	return listenConfig.Listen(ctx, network, address)
}

