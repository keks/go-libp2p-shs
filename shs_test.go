package shs

import (
	"context"
	"math/rand" // deterministic for tests
	"testing"

	b58 "github.com/jbenet/go-base58"
	shs "github.com/keks/secretstream/secrethandshake"
	transport "github.com/libp2p/go-libp2p-transport"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	appKeyLen   = 32
	testMessage = "Hello, World!"
)

var (
	t1, t2 *Transport
	k1, k2 *shs.EdKeyPair
	appKey [appKeyLen]byte
)

type randReader struct{}

func (r randReader) Read(buf []byte) (int, error) {
	return rand.Read(buf)
}

func init() {
	rand.Read(appKey[:])

	k1, _ = shs.GenEdKeyPair(randReader{})
	k2, _ = shs.GenEdKeyPair(randReader{})
}

func TestCreateTransport(t *testing.T) {
	t1 = NewTransport(*k1, appKey[:])
	t2 = NewTransport(*k2, appKey[:])
}

func TestConnect(t *testing.T) {
	// test transport interfaces
	var (
		tConn      transport.Conn
		tListener  transport.Listener
		tTransport transport.Transport
	)

	t1 = NewTransport(*k1, appKey[:])
	t2 = NewTransport(*k2, appKey[:])

	tTransport = t1 // compile time type check
	_ = tTransport

	laddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/1234/shs/" + b58.Encode(t1.keys.Public[:]))
	if err != nil {
		t.Fatal(err)
	}

	daddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/1235/shs/" + b58.Encode(t2.keys.Public[:]))
	if err != nil {
		t.Fatal(err)
	}

	l, err := t1.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}

	// compile time type checks
	tListener = l
	_ = tListener

	go func() {
		cs, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}

		// compile time type checks
		tConn = cs

		buf := make([]byte, 1024)
		n, err := cs.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if string(buf[:n]) != testMessage {
			t.Fatal("test message wrong:", string(buf[:n]))
		}
	}()

	d, err := t2.Dialer(daddr)
	if err != nil {
		t.Fatal(err)
	}

	cc, err := d.DialContext(context.WithValue(context.Background(), "t", t), laddr)
	if err != nil {
		t.Log("error dialing", laddr)
		t.Fatal(err)
	}

	n, err := cc.Write([]byte(testMessage))
	if err != nil {
		t.Fatal(err)
	}

	if n != len(testMessage) {
		t.Fatalf("c.Write returned %v, expected %v", n, len(testMessage))
	}
}
