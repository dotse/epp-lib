package epplib

import (
	"context"

	"github.com/beevik/etree"
)

// CommandFunc is the signature of a function which handles commands.
// The command xml information is in the etree.Document and the
// response should be written on the ResponseWriter.
// type HandlerFunc func(context.Context, *ResponseWriter, *etree.Document).
type CommandFunc func(context.Context, Writer, *etree.Document)

type handler struct {
	fn   CommandFunc
	path etree.Path
}
