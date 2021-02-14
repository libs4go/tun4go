package wc

import (
	"fmt"
	"strings"

	neturl "net/url"

	"github.com/libs4go/errors"
)

// URL wallet connect url
type URL struct {
	Topic   string `json:"topic"`   // handshake topic
	Version string `json:"version"` // Number (eg. 1.9.0)
	Bridge  string `json:"bridge"`  // Bridge URL (URL Encoded)
	Key     string `json:"key"`     // Symmetric key hex string
}

// ParseURL parse url string as URL object
func ParseURL(url string) (*URL, error) {

	if !strings.HasPrefix(url, "wc://") {
		url = strings.Replace(url, "wc:", "wc://", -1)
	}

	u, err := neturl.Parse(url)

	if err != nil {
		return nil, errors.Wrap(err, "parse %s error", url)
	}

	bridge := u.Query().Get("bridge")

	if bridge == "" {
		return nil, errors.Wrap(ErrURLBridge, "parse %s error", url)
	}

	key := u.Query().Get("key")

	if key == "" {
		return nil, errors.Wrap(ErrURLKey, "parse %s error", url)
	}

	return &URL{
		Topic:   u.User.Username(),
		Version: u.Host,
		Bridge:  bridge,
		Key:     key,
	}, nil
}

/// String implement Stringer
func (url *URL) String() string {
	return fmt.Sprintf("wc:%s@%s?bridge=%s&key=%s", url.Topic, url.Version, neturl.QueryEscape(url.Bridge), url.Key)
}
