package shs

import (
	"errors"
	"net"

	b58 "github.com/jbenet/go-base58"

	shs "github.com/keks/secretstream/secrethandshake"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	mafmt "github.com/whyrusleeping/mafmt"

	transport "github.com/libp2p/go-libp2p-transport"
)

var proto = ma.ProtocolWithName("shs")

var ErrTCPOnly = errors.New("shs only supports tcp") // TODO get rid of this
var ErrWrongBindKey = errors.New("public key in bind address doesn't match own key")

// Transport implements the go-libp2p-transport.Transport interface
type Transport struct {
	keys   shs.EdKeyPair
	appKey []byte
}

// NewTransport creates and initializes a struct of type *Transport
func NewTransport(k shs.EdKeyPair, appKey []byte) *Transport {
	return &Transport{k, appKey}
}

// Matches returns whether this Transport is capable of handling that Multiaddress
func (t *Transport) Matches(a ma.Multiaddr) bool {
	return mafmt.SHS.Matches(a)
}

// Dialer retuns a Dialer.
func (t *Transport) Dialer(laddr ma.Multiaddr, opts ...transport.DialOpt) (transport.Dialer, error) {
	nd := net.Dialer{}

	// set localaddr
	if laddr != nil {
		_, laTail := maHead(laddr)

		nladdr, err := manet.ToNetAddr(laTail)
		if err != nil {
			return nil, err
		}

		nd.LocalAddr = nladdr
	}

	return Dialer{nd, t}, nil
}

// Listen returns a *Listener for the specified laddr.
func (t *Transport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	head, tail := maHead(laddr)

	// get base58 pubkey from ma
	bindPubKey58, err := head.ValueForProtocol(proto.Code)
	if err != nil {
		return nil, err
	}

	// check if we bind to own pubkey
	if b58.Encode(t.keys.Public[:]) != bindPubKey58 {
		return nil, ErrWrongBindKey
	}

	// for the moment, until I found a better way TODO
	if _, err = tail.ValueForProtocol(ma.P_TCP); err != nil {
		return nil, ErrTCPOnly
	}

	// listen for incoming connections
	l, err := manet.Listen(tail)
	if err != nil {
		return nil, err
	}

	return &Listener{l, t}, nil
}
