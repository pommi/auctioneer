package listeners

import (
	"net"
	"time"
)

type ReadAheadConn struct {
	conn   net.Conn
	buffer []byte
}

func (r *ReadAheadConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (r *ReadAheadConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (r *ReadAheadConn) Close() error                       { return nil }
func (r *ReadAheadConn) LocalAddr() net.Addr                { return nil }
func (r *ReadAheadConn) RemoteAddr() net.Addr               { return nil }
func (r *ReadAheadConn) SetDeadline(t time.Time) error      { return nil }
func (r *ReadAheadConn) SetReadDeadline(t time.Time) error  { return nil }
func (r *ReadAheadConn) SetWriteDeadline(t time.Time) error { return nil }
