package tun4go

import (
	_ "github.com/google/uuid"       //
	_ "github.com/gorilla/websocket" //
	"github.com/libs4go/errors"
	_ "github.com/libs4go/errors"   //
	_ "github.com/libs4go/scf4go"   //
	_ "github.com/libs4go/slf4go"   //
	_ "github.com/stretchr/testify" //
)

// Params Tunnel initialize parameters
type Params map[string]string

// Tunnel tunnel protocol object
type Tunnel interface {
	// Send encrypt msg and send through provide transport
	Send(msg []byte, transport Transport) error

	// Recv recv msg through provide transport and decrypt
	Recv(transport Transport) ([]byte, error)

	// Disconnect send disconnect msg to peer
	Disconnect(transport Transport) error

	// Connect send connect msg to peer if need
	Connect(transport Transport) error

	// Get tunnel marshable context
	Context() ([]byte, error)
}

// TunnelCloser Tunnel object with close function
type TunnelCloser interface {
	Tunnel
	Close() error
}

// Transport underlying transport protocol
type Transport interface {
	Read() ([]byte, error)
	Write([]byte) error
}

// Approver .
type Approver interface {
	Approve(context []byte) bool
}

// Provider Tunnel protocol provider
type Provider interface {
	// Provider name
	Name() string

	// Create Tunnel from context bytes
	FromContext(context []byte) (Tunnel, error)

	// Create new tunnel with initialized context
	New(params Params) (Tunnel, error)
}

// New create new tunnel
func New(name string, params Params) (Tunnel, error) {
	var provider Provider

	if err := getProvider(name, &provider); err != nil {
		panic(errors.Wrap(err, "provider with name %s not found, call RegisterProvider first", name))
	}

	return provider.New(params)
}

// FromContext create tunnel with context
func FromContext(name string, context []byte) (Tunnel, error) {
	var provider Provider

	if err := getProvider(name, &provider); err != nil {
		panic(errors.Wrap(err, "provider with name %s not found, call RegisterProvider first", name))
	}

	return provider.FromContext(context)
}
