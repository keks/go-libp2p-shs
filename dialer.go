package shs

import (
	"context"

	"github.com/agl/ed25519"
	bs "github.com/keks/secretstream/boxstream"
	shs "github.com/keks/secretstream/secrethandshake"

	b58 "github.com/jbenet/go-base58"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"

	mafmt "github.com/whyrusleeping/mafmt"

	transport "github.com/libp2p/go-libp2p-transport"
)

// Dialer allows dialing other shs hosts
type Dialer struct {
	t *Transport
}

// Dial tries to connect to the shs host with the Multiaddress raddr.
func (d Dialer) Dial(raddr ma.Multiaddr) (transport.Conn, error) {
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

	state, err := shs.NewClientState(d.t.appKey, d.t.keys, rPub)
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
		remote:    state.Remote(),

		t: d.t,
	}

	return &boxStream, nil
}

// DialContext tries to connect to the shs host with the Multiaddress raddr.
func (d Dialer) DialContext(ctx context.Context, raddr ma.Multiaddr) (transport.Conn, error) {
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

	state, err := shs.NewClientState(d.t.appKey, d.t.keys, rPub)
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
		remote:    state.Remote(),

		t: d.t,
	}

	return &boxStream, nil
}

// Matches returns whether Multiaddress a can be reached using this Dialer.
func (d Dialer) Matches(a ma.Multiaddr) bool {
	return mafmt.SHS.Matches(a)
}
