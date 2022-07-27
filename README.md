# epp-lib

A collection of EPP nice to haves.

## namespace

All namespaces we support at the Swedish Internet Foundation are gathered here as
strings with some convenience methods.

## status

All statuses a response can have from the server
as described in `https://datatracker.ietf.org/doc/html/rfc5730#section-3`.

## error

Error represents an EPP error - it holds all the information
needed for an error response from the server.

## xml

Some nice to have convenience methods for xml.\
`XMLString` that automatically xml escape the given string when the
`Stringer` interface is used.\
`ParseXMLBool` function that handle `0`, `1`, `true` and `false`
and converts to go bool.\
`XMLPathBuilder` makes it easier to build xml paths.\
Example:

```go
// equals "name[namespace-uri()='urn:ietf:params:xml:ns:contact-1.0']"
NewXMLPathBuilder().AddOrphan("name", "urn:ietf:params:xml:ns:contact-1.0").String()

// equals "//command[namespace-uri()='random:namespace']/check[namespace-uri()='urn:ietf:params:xml:ns:contact-1.0']"
NewXMLPathBuilder().
  Add("//command", "random:namespace").
  Add("check", "urn:ietf:params:xml:ns:contact-1.0").String()
```

## server

A working EPP server with configurable variables like timeouts
and max message size allowed.\
Example server initialization:

```go
commandMux := &CommandMux{}

server := &Server{
    HandleCommand: commandMux.Handle,
    Greeting:      commandMux.GetGreeting,
    TLSConfig:     tls.Config{
            Certificates: []tls.Certificate{tlsCert},
            ClientAuth:   tls.RequireAnyClientCert,
            MinVersion: tls.VersionTLS12,
    },
    Timeout:        time.Hour,
    IdleTimeout:    350 * time.Second,
    WriteTimeout:   2 * time.Minute,
    ReadTimeout:    10 * time.Second,
    ErrorLog:       log, // whichever log you prefer that fit the interface
    MaxMessageSize: 1000,
}

listener, err := net.ListenTCP("tcp", tcpAddr)
if err != nil {
    panic(err)
}

if err := server.Serve(listener); err != nil {
    panic(err)
}
```

The server `ConnContext` can be used to set custom data on the context.
For example if you want to create a session ID for the connection or something.

The server `CloseConnHook` if set is called when a connection is closed.
It can be used to for example tear down any data for the connection.

## handler

The `CommandMux.Handle` function parse the incoming commands
and calls the correct `CommandFunc` if one has been
configured for the specific command.

Example bind of commands:

```go
commandMux := &CommandMux{}

commandMux.BindGreeting(funcThatHandlesGreetingCommand)
commandMux.Bind(
    xml.NewXMLPathBuilder().
        AddOrphan("//hello", NamespaceIETFEPP10.String()).String(),
    funcThatHandlesHelloCommand,
)
commandMux.BindCommand("info", NamespaceIETFContact10.String(),
    funcTharHandlesContactInfoCommand,
)

server.HandleCommand = commandMux.Handle
server.Greeting = commandMux.GetGreeting
```