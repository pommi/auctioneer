package listeners

import (
	"crypto/tls"
	"net"
)

//go:generate counterfeiter net.Listener
//go:generate counterfeiter net.Conn
//go:generate counterfeiter net.Addr

type upgradableTLSListener struct {
	listener  net.Listener
	tlsConfig *tls.Config
}

func NewUpgradableTLSListener(listener net.Listener) net.Listener {
	return &upgradableTLSListener{
		listener: listener,
	}
}

func (u *upgradableTLSListener) Accept() (net.Conn, error) {
	return &UpgradableConn{}, nil
}

func (u *upgradableTLSListener) Close() error {
	return u.listener.Close()
}

func (u *upgradableTLSListener) Addr() net.Addr {
	return u.listener.Addr()
}
