package shs

import (
	"github.com/agl/ed25519"
	bs "github.com/keks/secretstream/boxstream"
	shs "github.com/keks/secretstream/secrethandshake"

	b58 "github.com/jbenet/go-base58"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"

)

// Dialer allows dialing other shs hosts
type Dialer struct {
	keys   shs.EdKeyPair
	appKey []byte
}

// Dial tries to connect to the shs host with the Multiaddress raddr.
func (d Dialer) Dial(raddr ma.Multiaddr) (*Conn, error) {
	head, tail := maHead(raddr)

	rPubB58, err := head.ValueForProtocol(proto.Code)
	if err != nil {
		return nil, err
	}

	rPub := [ed25519.PublicKeySize]byte{}
	copy(rPub[:], b58.Decode(rPubB58))

	c, err := manet.Dial(tail)
	if err != nil {
		return nil, err
	}

	state, err := shs.NewClientState(d.appKey, d.keys, rPub)
	if err != nil {
		return nil, err
	}

	if err = shs.Client(state, c); err != nil {
		return nil, err
	}

	enKey, enNonce := state.GetBoxstreamEncKeys()
	deKey, deNonce := state.GetBoxstreamDecKeys()

	boxStream := Conn{
		Reader:    bs.NewUnboxer(c, &deNonce, &deKey),
		Writer:    bs.NewBoxer(c, &enNonce, &enKey),
		lowerConn: c,
		local:     d.keys.Public[:],
		remote:    state.Remote(),
	}

	return &boxStream, nil
}
