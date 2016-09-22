package shs

import (
	"net"
)

// netListener implements net.Listener so we can return it in Listener.NetListener
type netListener struct {
	l *Listener
}

func (nl *netListener) Accept() (net.Conn, error) {
	return nl.Accept()
}

func (nl *netListener) Close() error {
	return nl.Close()
}

func (nl *netListener) Addr() net.Addr {
	return nl.Addr()
}
