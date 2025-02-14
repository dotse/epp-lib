package epplib

import (
	"context"
	"io"
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
)

func TestMux_Greeting(t *testing.T) {
	t.Parallel()

	var greetingFuncCalled bool

	cm := &CommandMux{}
	cm.BindGreeting(func(ctx context.Context, writer Writer, document *etree.Document) {
		greetingFuncCalled = true
	})

	cm.GetGreeting(context.Background(), nil)

	assert.True(t, greetingFuncCalled)
}

func TestMux_Handle(t *testing.T) {
	t.Parallel()

	var (
		fooCalled bool
		barCalled bool
	)

	cm := &CommandMux{}

	cm.Bind(
		"//foo[namespace-uri()='urn:ietf:params:xml:ns:epp-1.0']",
		func(ctx context.Context, writer Writer, document *etree.Document) {
			fooCalled = true
		},
	)

	cm.Bind(
		"//bar[namespace-uri()='urn:ietf:params:xml:ns:epp-1.0']",
		func(ctx context.Context, writer Writer, document *etree.Document) {
			barCalled = true
		},
	)

	for _, tc := range []struct {
		name               string
		expectRwCloseAfter bool
		expectBarCalled    bool
		expectFooCalled    bool
		command            string
	}{
		{
			name:            "should call the correct func",
			expectFooCalled: true,
			command:         "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\n<epp xmlns=\"urn:ietf:params:xml:ns:epp-1.0\"\n     xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n     xsi:schemaLocation=\"urn:ietf:params:xml:ns:epp-1.0 epp-1.0.xsd\">\n<foo/>\n</epp>",
		},
		{
			name:               "should close read writer if command not found",
			expectRwCloseAfter: true,
			command:            "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\n<epp xmlns=\"urn:ietf:params:xml:ns:epp-1.0\"\n     xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n     xsi:schemaLocation=\"urn:ietf:params:xml:ns:epp-1.0 epp-1.0.xsd\">\n<test/>\n</epp>",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fooCalled = false
			barCalled = false

			rw := &ResponseWriter{}

			commandReader, commandWriter := io.Pipe()
			defer commandReader.Close()

			go writeAndClose(commandWriter, tc.command)

			cm.Handle(context.Background(), rw, commandReader)

			assert.Equal(t, tc.expectRwCloseAfter, rw.ShouldCloseAfterWrite())
			assert.Equal(t, tc.expectFooCalled, fooCalled)
			assert.Equal(t, tc.expectBarCalled, barCalled)
		})
	}
}

func TestMux_Bind(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		path       string
		handler    CommandFunc
		wantLength int
		wantPanic  bool
	}{
		{
			name:       "Binds to a valid path",
			path:       "//hello[namespace-uri()='urn:ietf:params:xml:ns:epp-1.0']",
			handler:    func(context.Context, Writer, *etree.Document) {},
			wantLength: 1,
			wantPanic:  false,
		},
		{
			name:      "Fails to bind with an incorrect path",
			path:      "[]",
			wantPanic: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cm := &CommandMux{}
			if tc.wantPanic {
				assert.Panics(t, func() {
					cm.Bind(tc.path, tc.handler)
				})
				return
			}
			cm.Bind(tc.path, tc.handler)
			assert.Equal(t, len(cm.handlers), tc.wantLength)
		})
	}
}

func TestMux_BindCommand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		ns         string
		handler    CommandFunc
		wantLength int
		wantPanic  bool
	}{
		{
			name:       "Binds to a valid path",
			ns:         NamespaceIETFHost10.String(),
			handler:    func(context.Context, Writer, *etree.Document) {},
			wantLength: 1,
			wantPanic:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cm := &CommandMux{}
			cm.BindCommand("test", tc.ns, tc.handler)
			assert.Equal(t, len(cm.handlers), tc.wantLength)
		})
	}
}

func writeAndClose(w io.WriteCloser, data string) {
	_, err := w.Write([]byte(data))
	if err != nil {
		panic(err)
	}

	w.Close()
}
