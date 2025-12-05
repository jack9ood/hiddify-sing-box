//go:build windows && go1.20

package dialer

import (
	"context"
	"io"
	"net"
	"os"
	"time"

	"github.com/sagernet/sing/common"
	M "github.com/sagernet/sing/common/metadata"
)

type slowOpenConn struct {
	dialer      net.Dialer
	ctx         context.Context
	network     string
	destination M.Socksaddr
	conn        net.Conn
}

func (c *slowOpenConn) Read(b []byte) (n int, err error) {
	if c.conn == nil {
		return 0, os.ErrInvalid
	}
	return c.conn.Read(b)
}

func (c *slowOpenConn) Write(b []byte) (n int, err error) {
	if c.conn != nil {
		return c.conn.Write(b)
	}
	// On Windows, we don't use TFO due to compatibility issues
	// Just establish a normal connection
	c.conn, err = c.dialer.DialContext(c.ctx, c.network, c.destination.String())
	if err != nil {
		return 0, err
	}
	return c.conn.Write(b)
}

func (c *slowOpenConn) Close() error {
	return common.Close(c.conn)
}

func (c *slowOpenConn) LocalAddr() net.Addr {
	if c.conn == nil {
		return M.Socksaddr{}
	}
	return c.conn.LocalAddr()
}

func (c *slowOpenConn) RemoteAddr() net.Addr {
	if c.conn == nil {
		return M.Socksaddr{}
	}
	return c.conn.RemoteAddr()
}

func (c *slowOpenConn) SetDeadline(t time.Time) error {
	if c.conn == nil {
		return os.ErrInvalid
	}
	return c.conn.SetDeadline(t)
}

func (c *slowOpenConn) SetReadDeadline(t time.Time) error {
	if c.conn == nil {
		return os.ErrInvalid
	}
	return c.conn.SetReadDeadline(t)
}

func (c *slowOpenConn) SetWriteDeadline(t time.Time) error {
	if c.conn == nil {
		return os.ErrInvalid
	}
	return c.conn.SetWriteDeadline(t)
}

func (c *slowOpenConn) ReaderReplaceable() bool {
	return c.conn == nil
}

func (c *slowOpenConn) WriterReplaceable() bool {
	return c.conn == nil
}

func (c *slowOpenConn) LazyHeadroom() bool {
	return c.conn == nil
}

func (c *slowOpenConn) Upstream() any {
	return c.conn
}

func (c *slowOpenConn) LazyWrite(payload []byte) error {
	_, err := c.Write(payload)
	return err
}

func (c *slowOpenConn) NeedHandshake() bool {
	return c.conn == nil
}

var _ io.Writer = (*slowWriteCloser)(nil)

type slowWriteCloser struct {
	writer io.Writer
	closer io.Closer
}

func (c *slowWriteCloser) Write(b []byte) (n int, err error) {
	return c.writer.Write(b)
}

func (c *slowWriteCloser) Close() error {
	return c.closer.Close()
}

