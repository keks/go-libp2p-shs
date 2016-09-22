package shs

import (
	"io"
	"net"
	"time"

	b58 "github.com/jbenet/go-base58"

	ma "github.com/jbenet/go-multiaddr"
	manet "github.com/jbenet/go-multiaddr-net"
)

// Conn is an encrypted connection to a remote shs host.
type Conn struct {
	io.Reader
	io.Writer

	lowerConn manet.Conn

	remote, local []byte
}

// LocalMultiaddr returns the local Multiaddr
func (c Conn) LocalMultiaddr() ma.Multiaddr {
	return ma.Join(c.lowerConn.LocalMultiaddr(), pubKeyToMA(c.local))
}

// LocalMultiaddr returns the remote end's Multiaddr
func (c Conn) RemoteMultiaddr() ma.Multiaddr {
	return ma.Join(c.lowerConn.RemoteMultiaddr(), pubKeyToMA(c.remote))
}

// Close closes the underlying net.Conn
func (c Conn) Close() error {
	return c.lowerConn.Close()
}

// LocalAddr returns the local net.Addr with the local public key
func (c Conn) LocalAddr() net.Addr {
	return Addr{c.lowerConn.LocalAddr(), c.local}
}

// RemoteAddr returns the remote net.Addr with the remote public key
func (c Conn) RemoteAddr() net.Addr {
	return Addr{c.lowerConn.RemoteAddr(), c.remote}
}

// SetDeadline passes the call to the underlying net.Conn
func (c Conn) SetDeadline(t time.Time) error {
	return c.lowerConn.SetDeadline(t)
}

// SetReadDeadline passes the call to the underlying net.Conn
func (c Conn) SetReadDeadline(t time.Time) error {
	return c.lowerConn.SetReadDeadline(t)
}

// SetWriteDeadline passes the call to the underlying net.Conn
func (c Conn) SetWriteDeadline(t time.Time) error {
	return c.lowerConn.SetWriteDeadline(t)
}

func pubKeyToMA(pub []byte) ma.Multiaddr {
	b58Str := b58.Encode(pub)

	a, err := ma.NewMultiaddr(ProtocolName + "/" + b58Str)
	if err != nil {
		panic(err) // TODO find a better way but interface doesn't accept errors
	}

	return a
}
