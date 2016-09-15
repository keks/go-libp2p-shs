package shs

import (
	b58 "github.com/jbenet/go-base58"

	ss "github.com/keks/go-libp2p-shs/thirdparty/secretstream"
	shs "github.com/keks/go-libp2p-shs/thirdparty/secretstream/secrethandshake"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

const (
	P_SHS        = 303
	ProtocolName = "shs"
)

var ErrTCPOnly = errors.New("shs only supports tcp") // TODO get rid of this

func init() {
	ma.AddProtocol(ma.Protocol{P_SHS, ma.LengthPrefixedVarSize, "shs", ma.CodeToVarint(P_SHS)})
}

// Transport implements the go-libp2p-transport.Transport interface
type Transport struct {
	keys   shs.EdKeyPair
	client ss.Client
}

// NewTransport creates and initializes a struct of type *Transport
func NewTransport(k shs.EdKeyPair, appKey []byte) *Transport {
	return &Transport{
		k,
		ss.NewClient(k, appKey),
	}
}

func (t *Transport) Dialer(laddr ma.Multiaddr, opts ...interface{}) (Dialer, error) {
	return Dialer(t.client), nil
}

func (t *Transport) Listen(laddr ma.Multiaddr) (*Listener, error) {
	head, tail := maHead(laddr)

	// get base58 pubkey from ma
	pubKey58, err := head.ValueForProtocol(P_SHS)
	if err != nil {
		return nil, err
	}

	// decode
	pubKey, err := b58.Decode(pubKey58)
	if err != nil {
		return nil, err
	}

	// check if its correct key TODO

	// for the moment, until I found a better way TODO
	if _, err = tail.ValueForCode(ma.P_TCP); err != nil {
		return nil, ErrTCPOnly
	}

	// listen for incoming connections
	l, err := manet.Listen(tail)
	if err != nil {
		return nil, err
	}

	return &Listener{l, t.keys}
}
