package epplib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// ErrMessageSize represent an error where the incoming message size is bigger than allowed.
var ErrMessageSize = errors.New("message size exceeds limit")

// MessageReader returns an io.Reader that reads one message according to the
// message size header. Blocks until a message size header is read.
// The message limit is in bytes.
func MessageReader(r io.Reader, msgLimit uint32) (io.Reader, error) {
	// Get the size of the message we are going to read in the future.
	// https://tools.ietf.org/html/rfc5734#section-4
	//
	// Note:The RFC doesn't explicitly state "signed" or "unsigned", but based
	// on its description the field is clearly meant to represent a
	// non-negative total length. Since a length can't be negative, it's best
	// to interpret the 32 bits as an unsigned integer, which is what we have
	// done historically.
	var totalSize uint32

	err := binary.Read(r, binary.BigEndian, &totalSize)
	if err != nil {
		return nil, err
	}

	// The length of the EPP XML instance is determined by subtracting four
	// octets from the total length of the data unit. The four octets are the
	// 32bit length field.
	if totalSize <= 4 || (msgLimit > 0 && totalSize > msgLimit) {
		return nil, fmt.Errorf("%w: incoming message size %d", ErrMessageSize, totalSize)
	}

	// Since we know the message size of the future message we can create a
	// reader that will read exactly that size and then return an EOF. That way
	// reading from the connection will always read the number of bytes that
	// the client said the message is.
	return io.LimitReader(r, int64(totalSize-4)), nil
}

// ResponseWriter is an io.Writer that buffers response data before writing it
// on the connection. Call CloseAfterWrite to close the connection after the
// response has been flushed.
type ResponseWriter struct {
	MessageBuffer
	closeAfterWrite bool
}

// CloseAfterWrite will set the flag to close the connection after the
// response has been flushed.
func (c *ResponseWriter) CloseAfterWrite() {
	c.closeAfterWrite = true
}

// ShouldCloseAfterWrite get to know if you should close after write.
func (c *ResponseWriter) ShouldCloseAfterWrite() bool {
	return c.closeAfterWrite
}

// MessageBuffer is a bytes.Buffer with a FlushTo method that will flush the
// contents of the buffer to a destination after writing the message size
// header.
type MessageBuffer struct {
	bytes.Buffer
}

// FlushTo flushes the buffer to dst after writing the message size header.
func (mb *MessageBuffer) FlushTo(dst io.Writer) error {
	// Begin by writing the len(b) as Big Endian uint32, including the
	// size of the content length header.
	// https://tools.ietf.org/html/rfc5734#section-4
	contentSize := mb.Len()
	if contentSize == 0 {
		// Nothing to write.
		return nil
	}

	totalSize := contentSize + 4 // 4 bytes for the content size header.
	if totalSize <= 4 || totalSize > math.MaxUint32 {
		return errors.New("invalid message size")
	}

	err := binary.Write(dst, binary.BigEndian, uint32(totalSize))
	if err != nil {
		return err
	}

	_, err = mb.WriteTo(dst)

	return err
}
