package shs

import (
	"errors"

	b58 "github.com/jbenet/go-base58"

	ma "github.com/jbenet/go-multiaddr"
	manet "github.com/jbenet/go-multiaddr-net"
	shs "github.com/keks/secretstream/secrethandshake"

const (
	ProtocolId   = 350
	ProtocolName = "shs"
)

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

func (t *Transport) Dialer(laddr ma.Multiaddr, opts ...interface{}) (Dialer, error) {
	return Dialer{t.keys, t.appKey}, nil
}

// Listen returns a *Listener for the specified laddr.
func (t *Transport) Listen(laddr ma.Multiaddr) (*Listener, error) {
	head, tail := maHead(laddr)

	// get base58 pubkey from ma
	bindPubKey58, err := head.ValueForProtocol(ProtocolId)
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

	return &Listener{l, t.keys, t.appKey}, nil
}
