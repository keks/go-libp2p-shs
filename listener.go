package shs

import (
	"net"

	b58 "github.com/jbenet/go-base58"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"

	bs "github.com/keks/secretstream/boxstream"
	shs "github.com/keks/secretstream/secrethandshake"

	transport "github.com/libp2p/go-libp2p-transport"
)

// Listener implements the go-libp2p-transport.Listener interface
type Listener struct {
	l manet.Listener
	t *Transport
}

// NetListener returns a net.Listener that is equivalent to this manet.Listener.
func (l *Listener) NetListener() net.Listener {
	return &netListener{l}
}

// Accept waits for an incoming connection and returns it. Else it returns an error.
func (l *Listener) Accept() (transport.Conn, error) {
	c, err := l.l.Accept()
	if err != nil {
		return nil, err
	}

	state, err := shs.NewServerState(l.t.appKey, l.t.keys)
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
		remote:    remote[:],

		t: l.t,
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
	return Addr{l.l.Addr(), l.t.keys.Public[:]}
}

// Multiaddr returns the Multiaddr the Listener bound to.
func (l *Listener) Multiaddr() ma.Multiaddr {
	return l.l.Multiaddr().Encapsulate(pubKeyToMA(l.t.keys.Public[:]))
}

// Addr implements net.Addr
type Addr struct {
	lower  net.Addr
	pubKey []byte
}

// Network returns the network we are on. Most likely "tcp/shs".
func (a Addr) Network() string {
	return a.lower.Network() + "/" + proto.Name
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
