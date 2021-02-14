package wc

import "github.com/libs4go/tun4go"

type wcProvider struct {
}

func newWCProvider() *wcProvider {
	return &wcProvider{}
}

func (provider *wcProvider) Name() string {
	return "wc"
}

func (provider *wcProvider) FromContext(context []byte) (tun4go.Tunnel, error) {
	return fromContext(context)
}

func (provider *wcProvider) New(params tun4go.Params) (tun4go.Tunnel, error) {
	return newWCTunnel(params)
}

func init() {
	tun4go.RegisterProvider(newWCProvider())
}
