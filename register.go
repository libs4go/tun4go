package tun4go

import (
	"fmt"
	"sync"

	"github.com/libs4go/sdi4go"
)

var injector sdi4go.Injector
var injectorOnce sync.Once

func getInjector() sdi4go.Injector {
	injectorOnce.Do(func() {
		injector = sdi4go.New()
	})

	return injector
}

// RegisterProvider .
func RegisterProvider(provider Provider) {
	getInjector().Bind(fmt.Sprintf("provider_%s", provider.Name()), sdi4go.Singleton(provider))
}

func getProvider(name string, objectPtr interface{}) error {
	return getInjector().Create(fmt.Sprintf("provider_%s", name), objectPtr)
}

func getEncoding(name string, objectPtr interface{}) error {
	return getInjector().Create(fmt.Sprintf("encoding_%s", name), objectPtr)
}
