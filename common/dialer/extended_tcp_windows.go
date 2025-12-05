//go:build windows && go1.20

package dialer

import (
	"context"
	"net"

	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

// Custom TCP dialer with extra features such as "TLS Fragmentation"
// TCP Fast Open is disabled on Windows due to compatibility issues
type ExtendedTCPDialer struct {
	net.Dialer
	DisableTFO  bool
	TLSFragment TLSFragment
}

func (d *ExtendedTCPDialer) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if !d.TLSFragment.Enabled || N.NetworkName(network) != N.NetworkTCP {
		switch N.NetworkName(network) {
		case N.NetworkTCP, N.NetworkUDP:
			return d.Dialer.DialContext(ctx, network, destination.String())
		default:
			return d.Dialer.DialContext(ctx, network, destination.AddrString())
		}
	}
	// Create a TLS-Fragmented dialer
	if d.TLSFragment.Enabled {
		fragmentConn := &fragmentConn{
			dialer:      d.Dialer,
			fragment:    d.TLSFragment,
			network:     network,
			destination: destination,
		}
		conn, err := d.Dialer.DialContext(ctx, network, destination.String())
		if err != nil {
			fragmentConn.err = err
			return nil, err
		}
		fragmentConn.conn = conn
		return fragmentConn, nil
	}
	// On Windows, TFO is disabled, use regular dialer wrapped in slowOpenConn for compatibility
	return &slowOpenConn{
			dialer:      d.Dialer,
			ctx:         ctx,
			network:     network,
			destination: destination,
		},
		nil
}

