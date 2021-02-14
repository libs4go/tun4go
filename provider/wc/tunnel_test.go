package wc

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/libs4go/errors"
	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/scf4go/reader/file"
	"github.com/libs4go/slf4go"
	_ "github.com/libs4go/slf4go/backend/console"
	"github.com/libs4go/tun4go"
	"github.com/stretchr/testify/require"
)

var url = "wc:fa854a0c-a88b-487d-8b25-88253b1a822f@1?bridge=https%3A%2F%2Fbridge.walletconnect.org&key=52661274329b1ec505c3ec29a870740145f7a9722b99b50f9c04c28164975ea5"

var transport *websockTransport

func init() {

	config := scf4go.New()

	err := config.Load(file.New(file.Yaml("./slf4go.yaml")))

	if err != nil {
		panic(err)
	}

	err = slf4go.Config(config)

	if err != nil {
		panic(err)
	}

	transport, err = newWebSockTransport(url)

	if err != nil {
		panic(err)
	}
}

type websockTransport struct {
	slf4go.Logger
	conn *websocket.Conn
}

func newWebSockTransport(url string) (*websockTransport, error) {
	u, err := ParseURL(url)

	if err != nil {
		return nil, errors.Wrap(err, "parse url %s error", url)
	}

	var bridge = u.Bridge

	if strings.HasPrefix(bridge, "http://") {
		bridge = "ws" + strings.TrimPrefix(bridge, "http")
	} else if strings.HasPrefix(bridge, "https://") {
		bridge = "wss" + strings.TrimPrefix(bridge, "https")
	}

	conn, _, err := websocket.DefaultDialer.Dial(bridge, nil)

	if err != nil {
		return nil, errors.Wrap(err, "dial to websocket server %s error", u.Bridge)
	}

	return &websockTransport{
		Logger: slf4go.Get("websocket"),
		conn:   conn,
	}, nil
}

func (trans *websockTransport) Read() ([]byte, error) {
	for {
		trans.D("Read message from bridge ...")
		t, message, err := trans.conn.ReadMessage()

		if err != nil {
			return nil, errors.Wrap(err, "read from websocket error")
		}

		if t != websocket.TextMessage {

			continue
		}

		trans.D("Recv msg: {@msg}", string(message))

		return message, nil

	}
}

func (trans *websockTransport) Write(buff []byte) error {

	trans.D("Send msg: {@msg}", string(buff))

	err := trans.conn.WriteMessage(websocket.TextMessage, buff)

	if err != nil {
		return errors.Wrap(err, "websocket send message error")
	}

	trans.D("Send msg -- success")

	return nil
}

func TestTunnel(t *testing.T) {

	defer slf4go.Sync()

	ci := marshal(&clientInfo{})

	tunnel, err := tun4go.New("wc", tun4go.Params{
		"clientinfo": ci,
		"account":    "0x120f18F5B8EdCaA3c083F9464c57C11D81a9E549",
		"url":        url,
	})

	require.NoError(t, err)

	err = tunnel.Connect(transport)

	require.NoError(t, err)

	buff, err := tunnel.Recv(transport)

	require.NoError(t, err)

	println(buff)
}

func marshal(v interface{}) string {
	buff, _ := json.Marshal(v)

	return string(buff)
}
