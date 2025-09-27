
package tor

import (
	"context"
	"net"
	"time"
)

// CheckSocks tries to connect to a SOCKS address to see if Tor is up.
func CheckSocks(ctx context.Context, addr string, timeout time.Duration) error {
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
