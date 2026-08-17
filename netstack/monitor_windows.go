package netstack

import (
	"context"
	"net"
)

// Windows uses a loopback UDP transport owned by the native helper. UDP has no
// portable peer-liveness probe; helper/child process supervision is authoritative.
func monitorTunnel(ctx context.Context, _ net.Conn) error {
	<-ctx.Done()
	return ctx.Err()
}
