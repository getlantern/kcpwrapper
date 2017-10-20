package kcpwrapper

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/getlantern/fdcount"
	"github.com/getlantern/keyman"
	"github.com/stretchr/testify/assert"
)

const (
	numClients = 100
)

func TestRoundTrip(t *testing.T) {
	_, fdc, err := fdcount.Matching("TCP")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		assert.NoError(t, fdc.AssertDelta(0))
	}()

	pk, err := keyman.GeneratePK(2048)
	if !assert.NoError(t, err) {
		return
	}
	cert, err := pk.TLSCertificateFor(time.Now().Add(365*24*time.Hour), true, nil, "kcpwrapper", "127.0.0.1")
	keypair, err := tls.X509KeyPair(cert.PEMEncoded(), pk.PEMEncoded())

	cfg := CommonConfig{
		DataShard:   10,
		Interval:    50,
		MTU:         1350,
		SockBuf:     4194304,
		ParityShard: 3,
		SndWnd:      128,
		Mode:        "fast2",
		Crypt:       "salsa20",
		Key:         "Lz7VwgTlp52dg9xWulDJlJzytniVerfSbOQdsAUPrnk=",
		RcvWnd:      512,
		KeepAlive:   10,
	}
	lcfg := &ListenerConfig{
		CommonConfig: cfg,
		Listen:       ":0",
	}
	dcfg := &DialerConfig{
		CommonConfig: cfg,
		Conn:         1,
		ScavengeTTL:  600,
	}

	_l, err := Listen(lcfg)
	if !assert.NoError(t, err) {
		return
	}
	l := tls.NewListener(_l, &tls.Config{
		Certificates: []tls.Certificate{keypair},
	})
	defer l.Close()

	go func() {
		for {
			conn, acceptErr := l.Accept()
			if acceptErr != nil {
				t.Logf("Unable to accept: %v", acceptErr)
				continue
			}
			go io.Copy(conn, conn)
		}
	}()

	resultCh := make(chan error, numClients)
	for i := 0; i < numClients; i++ {
		echoText := fmt.Sprintf("Hello Number %d", i)
		go func() {
			_conn, err := Dialer(dcfg)(context.Background(), "doesntmatter", l.Addr().String())
			if err != nil {
				resultCh <- err
			}
			conn := tls.Client(_conn, &tls.Config{
				ServerName: "127.0.0.1",
				RootCAs:    cert.PoolContainingCert(),
			})
			defer conn.Close()

			_, err = conn.Write([]byte(echoText))
			if err != nil {
				resultCh <- err
			}

			b := make([]byte, len(echoText))
			_, err = io.ReadFull(conn, b)
			if err != nil {
				resultCh <- err
			}

			if echoText != string(b) {
				resultCh <- fmt.Errorf("Mismatched echo text")
			}
			resultCh <- nil
		}()
	}

	for i := 0; i < numClients; i++ {
		assert.NoError(t, <-resultCh)
	}
}
