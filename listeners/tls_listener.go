package listeners

import (
	"crypto/tls"
	"errors"
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
	if u.listener == nil {
		return nil, errors.New("listener is nil")
	}

	c, err := u.listener.Accept()
	if err != nil {
		return nil, err
	}

	return &UpgradableConn{
		conn: c,
	}, nil
}

func (u *upgradableTLSListener) Close() error {
	return u.listener.Close()
}

func (u *upgradableTLSListener) Addr() net.Addr {
	return u.listener.Addr()
}
