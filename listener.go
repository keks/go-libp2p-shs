package shs

import (
	"net"

	ma "github.com/jbenet/go-multiaddr"
	manet "github.com/jbenet/go-multiaddr-net"
	ss "github.com/keks/go-libp2p-shs/thirdparty/secretstream"
	shs "github.com/keks/go-libp2p-shs/thirdparty/secretstream/secrethandshake"
)

// Listener implements the go-libp2p-transport.Listener interface
type Listener struct {
	l      manet.Listener
	keys   shs.EdKeyPair
	appKey []byte
}

func (l *Listener) Accept() (manet.Conn, error) {
	c, err := l.l.Accept()
	if err != nil {
		return nil, err
	}

	secConn, err := ss.ServerOnce(c, l.keys, l.appKey)
	return &Conn{secConn.(ss.Conn), c}, err
}

func (l *Listener) Close() error {
	// TODO maybe overwrite keys?
	return l.l.Close()
}

func (l *Listener) Addr() net.Addr {
	return Addr{l.l.Addr(), l.keys.Public[:]}
}

func (l *Listener) Multiaddr() ma.Multiaddr {
	return ma.Join(l.l.Multiaddr, pubKeyToMA(l.keys.Public[:]))
}

type Addr struct {
	lower  net.Addr
	pubKey []byte
}

func (a Addr) Network() string {
	return a.lower.Network() + "/" + ProtocolName
}

func (a Addr) PubKey() []byte {
	return a.pubKey
}

func (a Addr) String() string {
	return a.lower.String() + "/" + b58.Encode(a.pubKey)
}

func maHead(m ma.Multiaddr) (head, tail ma.Multiaddr) {
	ms := ma.Split(m)

	head = ms[len(ms)-1]
	tail = ma.Join(ms[:len(ms)-1])

	return
}
