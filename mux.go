package epplib

import (
	"context"
	"io"
	"log/slog"

	"github.com/beevik/etree"
)

// CommandMux parses and routes xml commands to bound handlers.
type CommandMux struct {
	greetingCommand CommandFunc
	handlers        []handler
}

// GetGreeting returns a greeting.
func (c *CommandMux) GetGreeting(ctx context.Context, rw *ResponseWriter) {
	c.greetingCommand(ctx, rw, nil)
}

// Handle handles a command. Commands will be routed according to how they are
// bound by the Bind function.
func (c *CommandMux) Handle(ctx context.Context, rw *ResponseWriter, cmd io.Reader) {
	doc := etree.NewDocument()

	_, err := doc.ReadFrom(cmd)
	if err != nil {
		slog.InfoContext(ctx, "could not read command",
			slog.Any("err", err),
		)

		rw.CloseAfterWrite()

		return
	}

	for _, h := range c.handlers {
		if el := doc.FindElementPath(h.path); el != nil {
			h.fn(ctx, rw, doc)
			return
		}
	}

	slog.InfoContext(ctx, "unknown command")
	rw.CloseAfterWrite()
}

// BindGreeting bind a greeting handler. Useful because EPP needs to send a
// greeting on connect.
func (c *CommandMux) BindGreeting(handler CommandFunc) {
	c.greetingCommand = handler
}

// Bind will bind a handler to a path.
func (c *CommandMux) Bind(path string, handlerFunc CommandFunc) {
	if c.handlers == nil {
		c.handlers = make([]handler, 0, 1)
	}

	c.handlers = append(c.handlers, handler{
		fn:   handlerFunc,
		path: etree.MustCompilePath(path),
	})
}

// BindCommand is a convenience method wrapping `Bind` with the common pattern used in
// epp. Note that it's currently hardcoded in the namespace-uri versions since there is
// currently only one.
func (c *CommandMux) BindCommand(command, ns string, handlerFunc CommandFunc) {
	c.Bind(NewXMLPathBuilder().
		AddOrphan("//command", "urn:ietf:params:xml:ns:epp-1.0").
		Add(command, "urn:ietf:params:xml:ns:epp-1.0").
		Add(command, ns).String(),
		handlerFunc,
	)
}
