package shs

import (
	"net"

	b58 "github.com/jbenet/go-base58"
	ma "github.com/jbenet/go-multiaddr"
	manet "github.com/jbenet/go-multiaddr-net"

	bs "github.com/keks/secretstream/boxstream"
	shs "github.com/keks/secretstream/secrethandshake"
)

// Listener implements the go-libp2p-transport.Listener interface
type Listener struct {
	l      manet.Listener
	keys   shs.EdKeyPair
	appKey []byte
}

// Accept waits for an incoming connection and returns it. Else it returns an error.
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

// Close closes the listener.
func (l *Listener) Close() error {
	// TODO maybe overwrite keys?
	return l.l.Close()
}

// Addr returns the net.Addr the Listener bound to.
func (l *Listener) Addr() net.Addr {
	return Addr{l.l.Addr(), l.keys.Public[:]}
}

// Multiaddr returns the Multiaddr the Listener bound to.
func (l *Listener) Multiaddr() ma.Multiaddr {
	return l.l.Multiaddr().Encapsulate(pubKeyToMA(l.keys.Public[:]))
}

// Addr implements net.Addr
type Addr struct {
	lower  net.Addr
	pubKey []byte
}

// Network returns the network we are on. Most likely "tcp/shs".
func (a Addr) Network() string {
	return a.lower.Network() + "/" + ProtocolName
}

// PubKey returns the public key of the node at this address
func (a Addr) PubKey() []byte {
	return a.pubKey
}

// String is the string representation of this address.
func (a Addr) String() string {
	return a.lower.String() + "/" + b58.Encode(a.pubKey)
}

func maHead(m ma.Multiaddr) (head, tail ma.Multiaddr) {
	ms := ma.Split(m)

	head = ms[len(ms)-1]
	tail = ma.Join(ms[:len(ms)-1]...)

	return
}
