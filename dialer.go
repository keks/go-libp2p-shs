package shs

import (
	b58 "github.com/jbenet/go-base58"

	shs "github.com/keks/go-libp2p-shs/thirdparty/secretstream/secrethandshake"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

type Dialer struct {
	keys   shs.EdKeyPair
	appKey []byte
}

func (d Dialer) Dial(raddr ma.Multiaddr) (*Conn, error) {
	head, tail := maHead(raddr)

	rPubB58, err := head.ValueForProtocol(P_SHS)
	if err != nil {
		return nil, err
	}

	rPub := [ed25519.PublicKeySize]byte{}
	copy(rPub[:], b58.Decode(rPub58))

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
		Reader: bs.NewUnboxer(conn, &deNonce, &deKey),
		Writer: bs.NewBoxer(conn, &enNonce, &enKey),
		conn:   c,
		local:  c.keys.Public[:],
		remote: state.Remote(),
	}

	return boxStream, nil
}
