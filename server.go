package epplib

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Server is an EPP server.
type Server struct {
	// HandleCommand handles commands for a connection. Write the response on rw and
	// read the command from cmd.
	HandleCommand func(ctx context.Context, rw *ResponseWriter, cmd io.Reader)

	// Greeting is called once when a new connection is established and should
	// write a greeting on rw.
	Greeting func(ctx context.Context, rw *ResponseWriter)

	// ConnContext can be used to add metadata to a connection. All future
	// calls to the Handler will include this context.
	ConnContext func(ctx context.Context, conn *tls.Conn) (context.Context, error)

	// CloseConnHook can be used to run a function after a connection has been closed,
	// taking the closed connection as an argument.
	CloseConnHook func(ctx context.Context, conn *tls.Conn)

	TLSConfig tls.Config

	// Timeout is the total time a connection can stay open on the server.
	// After this duration the connection is automatically closed.
	Timeout time.Duration

	// IdleTimeout is how long the connection will stay open without any
	// activity.
	IdleTimeout time.Duration

	// WriteTimeout is how long to wait for writes on the response writer.
	WriteTimeout time.Duration

	// ReadTimeout is how long to wait for reading a command.
	ReadTimeout time.Duration

	// MaxMessageSize if set will return an error if the incoming request
	// is bigger than the set size in bytes. 0 indicates no limit.
	MaxMessageSize int64

	// Logger logs errors when accepting connections, unexpected behavior
	// from handlers and underlying connection errors.
	Logger Logger

	// We keep track of our active connections here. This is guarded by mu.
	activeConn map[*eppConn]struct{}

	// mu guards activeConn.
	mu sync.Mutex

	// counts active connections, in created on new connections and decreased
	// when connections are closed.
	wg sync.WaitGroup

	// where we accept new connections.
	listener   Listener
	listenerMu sync.RWMutex
}

// Serve will start a server on the provided listener.
func (s *Server) Serve(listener Listener) error {
	if s.HandleCommand == nil || s.Greeting == nil {
		panic("Handler and Greeting is required")
	}

	s.listenerMu.Lock()
	s.listener = listener
	s.listenerMu.Unlock()

	s.activeConn = make(map[*eppConn]struct{})

	defer func() {
		s.mu.Lock()

		for c := range s.activeConn {
			_ = c.stopAwaitMessage()
		}

		s.mu.Unlock()

		s.wg.Wait()
	}()

	for {
		// mutex for listener to not accept new connections when closing the server
		s.listenerMu.RLock()
		conn, err := s.listener.Accept()
		s.listenerMu.RUnlock()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				// The listener was closed, this happens when we stop the
				// server and shouldn't count as an error.
				return nil
			}

			var netErr net.Error

			if errors.As(err, &netErr) && netErr.Timeout() {
				// Any temporary errors can be retried so we can just continue
				// here.
				continue
			}

			// Unexpected errors should stop the server.
			return err
		}

		keepAliveConn, ok := conn.(KeepAliveConn)
		if !ok {
			return errors.New("connection does not support keep alive")
		}

		if err := keepAliveConn.SetKeepAlive(true); err != nil {
			return err
		}

		if err := keepAliveConn.SetKeepAlivePeriod(time.Minute); err != nil {
			return err
		}

		s.wg.Add(1)

		go s.serveConn(conn)
	}
}

// Close will gracefully stop the server.
func (s *Server) Close() error {
	s.listenerMu.RLock()
	err := s.listener.Close()
	s.listenerMu.RUnlock()

	return err
}

func (s *Server) serveConn(conn net.Conn) {
	tlsConn := tls.Server(conn, s.TLSConfig.Clone())

	c := &eppConn{conn: tlsConn, maxMessageSize: s.MaxMessageSize}

	s.mu.Lock()
	s.activeConn[c] = struct{}{}
	s.mu.Unlock()

	// Set up a cancel context that is passed to handlers so that they, if needed,
	// can be notified when the connection shuts down.
	ctx, cancelCtx := context.WithCancel(context.Background())

	// Setup some cleanup for when the session exits.
	defer func() {
		_ = c.Close()

		// No need to remember the closeChan anymore.
		s.mu.Lock()
		delete(s.activeConn, c)
		s.mu.Unlock()

		if s.CloseConnHook != nil {
			s.CloseConnHook(ctx, c.conn.(*tls.Conn))
		}

		cancelCtx()

		// Countdown the wait group so that the entire listener can shut down
		// when this reaches zero if it wants to.
		s.wg.Done()
	}()

	err := setDeadlines(c.conn, s.ReadTimeout, s.WriteTimeout)
	if err != nil {
		s.logError("handshake deadlines", err)
		return
	}

	err = tlsConn.Handshake()
	if err != nil {
		s.logDebug("handshake", err)
		return
	}

	if s.ConnContext != nil {
		// This is where the user can set up any context data for the
		// connection, for example userID's etc.
		ctx, err = s.ConnContext(ctx, tlsConn)
		if err != nil {
			// We don't log the error here because it's only here to signal
			// that the user wanted to abort the connection.
			return
		}
	}

	// The responseWriter can be reused for each command.
	rw := ResponseWriter{}

	err = setDeadlines(c.conn, s.ReadTimeout, s.WriteTimeout)
	if err != nil {
		s.logError("set deadlines for greeting", err)
		return
	}

	// We have properly connected so we need to begin by sending the greeting.
	s.Greeting(ctx, &rw)

	err = rw.FlushTo(c.conn)
	if err != nil {
		s.logError("flush greeting", err)
		return
	}

	if rw.ShouldCloseAfterWrite() {
		return
	}

	maxDeadline := deadlineFromTimeout(s.Timeout)

	for {
		deadline := getClosestDeadline(
			maxDeadline,
			deadlineFromTimeout(s.IdleTimeout),
		)

		err := c.conn.SetDeadline(deadline)
		if err != nil {
			s.logError("set deadlines for await message", err)
			return
		}

		// Wait for a message to appear on the connection.
		cmd, err := c.AwaitMessage()
		if err != nil { // nolint:nestif // Needed.
			if errors.Is(err, os.ErrDeadlineExceeded) {
				// We have reached the deadline for this session, we now need
				// to disconnect.
				return
			}

			if errors.Is(err, net.ErrClosed) ||
				errors.Is(err, syscall.ECONNRESET) ||
				errors.Is(err, syscall.EPIPE) {
				// Either we have closed AwaitMessage or the client has closed
				// the connection.
				return
			}

			if errors.Is(err, io.EOF) {
				// The client has closed the connection.
				return
			}

			if errors.Is(err, io.ErrUnexpectedEOF) {
				// We don't want to turn this of entirely
				s.logInfo("await message", err)
				return
			}

			if errors.Is(err, ErrMessageSize) {
				// Client has told us that the incoming message is larger than
				// our supported max size of a message.
				s.logInfo(fmt.Sprintf("Message limit exceeded from %q", c.conn.RemoteAddr()), err)

				return
			}

			if strings.Contains(err.Error(), "user canceled") {
				// From RFC 5247: https://datatracker.ietf.org/doc/html/rfc5246#section-7.2.2
				// user_canceled: This handshake is being canceled for some reason unrelated to a
				// protocol failure.  If the user cancels an operation after the
				// handshake is complete, just closing the connection by sending a
				// close_notify is more appropriate.  This alert should be followed
				// by a close_notify.  This message is generally a warning.
				s.logInfo(fmt.Sprintf("handshake was canceled by client %q", c.conn.RemoteAddr()), err)
				return
			}

			// We have some other error
			s.logError(fmt.Sprintf("await message from %q", c.conn.RemoteAddr()), err)

			return
		}

		err = setDeadlines(c.conn, s.ReadTimeout, s.WriteTimeout)
		if err != nil {
			s.logError("set deadlines for command read/write", err)
			return
		}

		// We have some command that is waiting to be read.
		s.HandleCommand(ctx, &rw, cmd)

		// Flush the message to the underlying connection.
		err = rw.FlushTo(c.conn)
		if err != nil {
			if errors.Is(err, syscall.EPIPE) {
				// The client has closed the connection. I.e. "broken pipe".
				s.logInfo("flush response", err)
				return
			}

			if strings.Contains(err.Error(), "connection reset by peer") {
				// Client has closed the connection with a TCP RST package.
				// Wierdly enough this is not caught by a errors.Is(err, sycall.ECONNRESET)
				// like is done above in awaitMessage. Therefore we will info log here in-
				// case catch to broadly here.
				s.logInfo("flush response", err)
				return
			}

			s.logError("flush response", err)

			return
		}

		if rw.ShouldCloseAfterWrite() {
			return
		}
	}
}

// CloseConnection will gracefully close the provided conn.
func (s *Server) CloseConnection(conn *tls.Conn) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for c := range s.activeConn {
		if c.conn == conn {
			err := c.stopAwaitMessage()
			if err != nil {
				return c.Close()
			}
		}
	}

	return nil
}

func (s *Server) logError(prefix string, err error) {
	if s.Logger != nil {
		s.Logger.Errorf("epp: %s: %+v", prefix, err)
	}
}

func (s *Server) logInfo(prefix string, err error) {
	if s.Logger != nil {
		s.Logger.Infof("epp: %s: %+v", prefix, err)
	}
}

func (s *Server) logDebug(prefix string, err error) {
	if s.Logger != nil {
		s.Logger.Debugf("epp: %s: %v", prefix, err)
	}
}

func setDeadlines(conn net.Conn, readTimeout, writeTimeout time.Duration) error {
	err := conn.SetReadDeadline(deadlineFromTimeout(readTimeout))
	if err != nil {
		return err
	}

	err = conn.SetWriteDeadline(deadlineFromTimeout(writeTimeout))
	if err != nil {
		return err
	}

	return nil
}

// getClosestDeadline return the deadline that is closest to the current time.
func getClosestDeadline(dls ...time.Time) time.Time {
	closest := time.Time{}
	now := time.Now()

	for _, dl := range dls {
		if dl.IsZero() || dl.Before(now) {
			// Skip dls that are before the current time.
			continue
		}

		if closest.After(dl) || closest.IsZero() {
			closest = dl
		}
	}

	return closest
}

func deadlineFromTimeout(timeout time.Duration) time.Time {
	if timeout == 0 {
		return time.Time{}
	}

	return time.Now().Add(timeout)
}

type eppConn struct {
	conn net.Conn

	// isAwaitingMsg is 1 while we are waiting for a size header.
	isAwaitingMsg int32

	// stopAwaitMsg is 1 when we no longer want to wait for size headers. This
	// causes AwaitMessage to return net.ErrClosed immediately when called.
	stopAwaitMsg int32

	maxMessageSize int64
}

// AwaitMessage blocks until a message header is read from the underlying
// connection. After this is called Read will return EOF when the entire
// message is read.
func (c *eppConn) AwaitMessage() (io.Reader, error) {
	// Remember that we are awaiting a message. This needs to be done before we
	// check if we even should await more messages so that we always set the
	// deadline correctly in the stopAwaitMessage function.
	atomic.StoreInt32(&c.isAwaitingMsg, 1)

	if atomic.LoadInt32(&c.stopAwaitMsg) == 1 {
		// We shouldn't await messages at all, this function is "closed".
		atomic.StoreInt32(&c.isAwaitingMsg, 0)
		return nil, net.ErrClosed
	}

	msgReader, err := MessageReader(c.conn, c.maxMessageSize)

	// We are no longer awaiting. We have a message to be processed.
	// We set this to 0 before checking any errors.
	atomic.StoreInt32(&c.isAwaitingMsg, 0)

	return msgReader, err
}

// stopAwaitMessage will unblock and close the AwaitMessage function. Future
// calls to AwaitMessage will return net.ErrClosed.
func (c *eppConn) stopAwaitMessage() error {
	atomic.StoreInt32(&c.stopAwaitMsg, 1)

	if atomic.LoadInt32(&c.isAwaitingMsg) == 1 {
		// We're currently awaiting a message, interrupt it without closing the
		// connection.
		return c.conn.SetDeadline(time.Now())
	}

	return nil
}

// Close will close the underlying connection.
func (c *eppConn) Close() error {
	return c.conn.Close()
}
