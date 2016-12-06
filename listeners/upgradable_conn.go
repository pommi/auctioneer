package listeners

import (
	"crypto/tls"
	"net"
	"time"
)

type UpgradableConn struct {
	conn         net.Conn
	tlsConfig    *tls.Config
	iniitalizing bool
}

func (u *UpgradableConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (u *UpgradableConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (u *UpgradableConn) Close() error                       { return nil }
func (u *UpgradableConn) LocalAddr() net.Addr                { return nil }
func (u *UpgradableConn) RemoteAddr() net.Addr               { return nil }
func (u *UpgradableConn) SetDeadline(t time.Time) error      { return nil }
func (u *UpgradableConn) SetReadDeadline(t time.Time) error  { return nil }
func (u *UpgradableConn) SetWriteDeadline(t time.Time) error { return nil }
