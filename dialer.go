package shs

import (
	"context"
	"net"
	"testing"

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
	net.Dialer

	t *Transport
}

func (d Dialer) Dial(raddr ma.Multiaddr) (transport.Conn, error) {
	return d.DialContext(context.Background(), raddr)
}

// DialContext tries to connect to the shs host with the Multiaddress raddr.
func (d Dialer) DialContext(ctx context.Context, raddr ma.Multiaddr) (transport.Conn, error) {
	var t *testing.T = nil
	if t_, ok := ctx.Value("t").(*testing.T); ok {
		t = t_
	}

	head, tail := maHead(raddr)

	rPubB58, err := head.ValueForProtocol(proto.Code)
	if err != nil {
		return nil, err
	}

	rPub := [ed25519.PublicKeySize]byte{}
	copy(rPub[:], b58.Decode(rPubB58))

	nraddr, err := manet.ToNetAddr(tail)
	if err != nil {
		return nil, err
	}

	c, err := d.Dialer.DialContext(ctx, "tcp", nraddr.String())
	if err != nil {
		if t != nil {
			t.Log("bla")
		}
		return nil, err
	}

	state, err := shs.NewClientState(d.t.appKey, d.t.keys, rPub)
	if err != nil {
		if t != nil {
			t.Log("bla")
		}
		return nil, err
	}

	if err = shs.Client(state, c); err != nil {
		if t != nil {
			t.Log("bla")
		}
		return nil, err
	}

	manetConn, err := manet.WrapNetConn(c)
	if err != nil {
		if t != nil {
			t.Log("bla")
		}
		return nil, err
	}

	enKey, enNonce := state.GetBoxstreamEncKeys()
	deKey, deNonce := state.GetBoxstreamDecKeys()

	boxStream := Conn{
		Reader:    bs.NewUnboxer(c, &deNonce, &deKey),
		Writer:    bs.NewBoxer(c, &enNonce, &enKey),
		lowerConn: manetConn,
		remote:    state.Remote(),

		t: d.t,
	}

	return &boxStream, nil
}

// Matches returns whether Multiaddress a can be reached using this Dialer.
func (d Dialer) Matches(a ma.Multiaddr) bool {
	return mafmt.SHS.Matches(a)
}
