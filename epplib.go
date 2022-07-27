package epplib

import (
	"io"
	"net"
	"time"
)

// Writer can write and close writers.
type Writer interface {
	io.Writer
	CloseAfterWrite()
	Reset()
}

// Listener can accept new connections.
type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
}

// KeepAliveConn can set keep alive information.
type KeepAliveConn interface {
	SetKeepAlive(b bool) error
	SetKeepAlivePeriod(d time.Duration) error
}
