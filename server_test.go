package epplib

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenAndServe(t *testing.T) {
	t.Parallel()

	s := Server{
		HandleCommand: func(ctx context.Context, rw *ResponseWriter, cmd io.Reader) {},
		Greeting:      func(ctx context.Context, rw *ResponseWriter) {},
	}
	defer s.Close()

	go func() {
		tcpAddr, err := net.ResolveTCPAddr("tcp", ":")
		require.NoError(t, err)

		listener, err := net.ListenTCP("tcp", tcpAddr)
		require.NoError(t, err)

		err = s.Serve(listener)
		require.NoError(t, err)
	}()

	client := dialServer(t, &s, &tls.Config{InsecureSkipVerify: true})
	err := client.Handshake()
	require.Contains(t, err.Error(), "remote error: tls")
}

func TestListenAndServeClose(t *testing.T) {
	t.Parallel()

	s := Server{
		HandleCommand: func(ctx context.Context, rw *ResponseWriter, cmd io.Reader) {},
		Greeting:      func(ctx context.Context, rw *ResponseWriter) {},
		TLSConfig: tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{generateCertificate()},
		},
	}

	closed := make(chan struct{})

	go func() {
		tcpAddr, err := net.ResolveTCPAddr("tcp", ":")
		require.NoError(t, err)

		listener, err := net.ListenTCP("tcp", tcpAddr)
		require.NoError(t, err)

		err = s.Serve(listener)
		require.NoError(t, err)

		close(closed)
	}()

	// Dial it, so we know it's started and listening.
	require.NoError(t, dialServer(t, &s, &tls.Config{InsecureSkipVerify: true}).Handshake())

	err := s.Close()
	require.NoError(t, err)

	select {
	case <-closed:
	case <-time.After(10 * time.Second):
		t.Fatal("listener was not closed")
	}
}

func TestCloseConnHookCalled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		wantCloseHookCalled bool
		closeConnHook       func(context.Context, *tls.Conn)
	}{
		{
			name:                "Calls the close conn hook function",
			wantCloseHookCalled: true,
			closeConnHook: func(ctx context.Context, conn *tls.Conn) {
				assert.NotNil(t, conn)

				b := make([]byte, 10)
				_, err := conn.Read(b) // Make sure that the conn is closed
				require.Error(t, err)
			},
		},
		{
			name:                "Does not call close conn hook function",
			wantCloseHookCalled: false,
		},
	}

	for _, tt := range cases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				clientConn, serverConn = net.Pipe()
				clientTLSConn          = tls.Client(clientConn, &tls.Config{InsecureSkipVerify: true})
				closeConnHook          = tt.closeConnHook
				closeConnHookCalled    bool
			)

			if tt.wantCloseHookCalled {
				closeConnHook = func(ctx context.Context, conn *tls.Conn) {
					closeConnHookCalled = tt.wantCloseHookCalled

					if tt.closeConnHook != nil {
						tt.closeConnHook(ctx, conn)
					}
				}
			}

			s := Server{
				Greeting:      func(ctx context.Context, rw *ResponseWriter) {},
				CloseConnHook: closeConnHook,
				IdleTimeout:   100 * time.Millisecond,
				TLSConfig: tls.Config{
					InsecureSkipVerify: true,
					Certificates:       []tls.Certificate{generateCertificate()},
				},
				activeConn: make(map[*eppConn]struct{}),
			}

			s.wg.Add(1)

			go s.serveConn(serverConn)

			err := clientTLSConn.Handshake()
			require.NoError(t, err)

			assert.NoError(t, clientTLSConn.Close())

			s.wg.Wait()

			if !tt.wantCloseHookCalled {
				assert.False(t, closeConnHookCalled)
				return
			}

			assert.True(t, closeConnHookCalled)
		})
	}
}

func TestGreeting(t *testing.T) {
	t.Parallel()

	s := Server{
		TLSConfig: tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{generateCertificate()},
		},
		activeConn: make(map[*eppConn]struct{}),
		Greeting: func(ctx context.Context, rw *ResponseWriter) {
			_, err := fmt.Fprint(rw, "Greeting")
			assert.NoError(t, err)
		},
	}

	clientConn, serverConn := net.Pipe()

	go s.serveConn(serverConn)

	clientTLSConn := tls.Client(clientConn, &tls.Config{InsecureSkipVerify: true})
	err := clientTLSConn.Handshake()
	require.NoError(t, err)

	msg := getMessage(t, clientTLSConn)
	assert.Equal(t, "Greeting", msg)
}

func TestSendReceiveMessage(t *testing.T) {
	t.Parallel()

	s := Server{
		TLSConfig: tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{generateCertificate()},
		},
		activeConn: make(map[*eppConn]struct{}),
		Greeting: func(ctx context.Context, rw *ResponseWriter) {
			_, err := fmt.Fprint(rw, "Greeting")
			assert.NoError(t, err)
		},
		HandleCommand: func(ctx context.Context, rw *ResponseWriter, cmd io.Reader) {
			data, _ := io.ReadAll(cmd)
			_, err := fmt.Fprintf(rw, "Response to: %s", string(data))
			assert.NoError(t, err)
		},
		IdleTimeout: 10 * time.Second,
	}

	clientConn, serverConn := net.Pipe()

	s.wg.Add(1)

	go s.serveConn(serverConn)

	clientTLSConn := tls.Client(clientConn, &tls.Config{InsecureSkipVerify: true})
	err := clientTLSConn.Handshake()
	require.NoError(t, err)

	msg := getMessage(t, clientTLSConn)
	assert.Equal(t, "Greeting", msg)

	buf := MessageBuffer{}

	// Write a command.
	_, err = buf.WriteString("A command")
	assert.NoError(t, err)
	assert.NoError(t, buf.FlushTo(clientTLSConn))

	// We should have a response.
	resp := getMessage(t, clientTLSConn)
	assert.Equal(t, "Response to: A command", resp)
}

func TestReceiveTooBigMessage(t *testing.T) {
	t.Parallel()

	s := Server{
		TLSConfig: tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{generateCertificate()},
		},
		activeConn: make(map[*eppConn]struct{}),
		Greeting: func(ctx context.Context, rw *ResponseWriter) {
			b := make([]byte, 10)
			_, err := rand.Read(b)
			require.NoError(t, err)
			_, err = fmt.Fprint(rw, b)
			assert.NoError(t, err)
		},
		IdleTimeout: 10 * time.Second,
	}

	clientConn, serverConn := net.Pipe()

	s.wg.Add(1)

	go s.serveConn(serverConn)

	clientTLSConn := tls.Client(clientConn, &tls.Config{InsecureSkipVerify: true})
	err := clientTLSConn.Handshake()
	require.NoError(t, err)

	_, err = MessageReader(clientTLSConn, 9)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrMessageSize))
}

func getMessage(t *testing.T, r io.Reader) string {
	msgReader, err := MessageReader(r, 0)
	require.NoError(t, err)

	data, err := io.ReadAll(msgReader)
	require.NoError(t, err)

	return string(data)
}

func dialServer(t *testing.T, s *Server, config *tls.Config) *tls.Conn {
	var (
		conn net.Conn
		err  error
	)

	for i := 0; i < 100; i++ {
		time.Sleep(time.Duration(i) * time.Millisecond)

		addr := ""

		s.listenerMu.RLock()
		if s.listener != nil {
			addr = s.listener.Addr().String()
		}
		s.listenerMu.RUnlock()

		if addr == "" {
			continue
		}

		conn, err = net.Dial("tcp", addr)
		if err != nil {
			continue
		}

		break
	}

	require.NoError(t, err)

	return tls.Client(conn, config)
}

func generateCertificate() tls.Certificate {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			CommonName:   "epp.example.test",
			Organization: []string{"Simple Server Test"},
			Country:      []string{"SE"},
			Locality:     []string{"Stockholm"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	certificate, _ := x509.CreateCertificate(rand.Reader, cert, cert, key.Public(), key)

	return tls.Certificate{
		Certificate: [][]byte{certificate},
		PrivateKey:  key,
	}
}
