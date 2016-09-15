package shs

import (
	"net"

	b58 "github.com/jbenet/go-base58"
	ma "github.com/jbenet/go-multiaddr"
	manet "github.com/jbenet/go-multiaddr-net"

	bs "github.com/keks/go-libp2p-shs/thirdparty/secretstream/boxstream"
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

	state, err := shs.NewServerState(l.appKey, l.keys)
	if err != nil {
		return nil, err
	}

	err = shs.Server(state, c)
	if err != nil {
		return nil, err
	}

	enKey, enNonce := state.GetBoxstreamEncKeys()
	deKey, deNonce := state.GetBoxstreamDecKeys()

	remote := state.Remote()
	boxed := Conn{
		Reader:    bs.NewUnboxer(c, &deNonce, &deKey),
		Writer:    bs.NewBoxer(c, &enNonce, &enKey),
		lowerConn: c,
		local:     l.keys.Public[:],
		remote:    remote[:],
	}

	return boxed, nil
}

func (l *Listener) Close() error {
	// TODO maybe overwrite keys?
	return l.l.Close()
}

func (l *Listener) Addr() net.Addr {
	return Addr{l.l.Addr(), l.keys.Public[:]}
}

func (l *Listener) Multiaddr() ma.Multiaddr {
	return l.l.Multiaddr().Encapsulate(pubKeyToMA(l.keys.Public[:]))
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
	tail = ma.Join(ms[:len(ms)-1]...)

	return
}
