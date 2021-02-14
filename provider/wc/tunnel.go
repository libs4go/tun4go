package wc

import (
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/libs4go/errors"
	"github.com/libs4go/slf4go"
	"github.com/libs4go/tun4go"
)

// Status Tunnel status
type Status string

// Status enum
const (
	Connecting    Status = "connecting"
	Connected     Status = "connected"
	Disconnecting Status = "disconnecting"
	Disconnected  Status = "disconnected"
)

type clientInfo struct {
	Description string   `json:"description"`
	URL         string   `json:"url,omitempty"`
	ICONs       []string `json:"icons,omitempty"`
	Name        string   `json:"name"`
}

type wcTunnel struct {
	slf4go.Logger `json:"-"`
	URL           *URL        `json:"url"`
	Self          string      `json:"self"`
	SelfInfo      *clientInfo `json:"self-info"`
	Key           []byte      `json:"key"`
	PeerInfo      *clientInfo `json:"peer-info"`
	Peer          string      `json:"peer"`
	ChainID       int64       `json:"chain-id"`
	Accounts      []string    `json:"accounts"`
	Status        Status      `json:"status"`
}

func newWCTunnel(params tun4go.Params) (*wcTunnel, error) {

	url, ok := params["url"]

	if !ok {
		return nil, errors.Wrap(ErrParams, "expect handshake url param")
	}

	u, err := ParseURL(url)

	if err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(u.Key)

	if err != nil {
		return nil, errors.Wrap(err, "decode key %s error", u.Key)
	}

	account, ok := params["account"]

	if !ok {
		return nil, errors.Wrap(ErrParams, "expect account param")
	}

	var ci *clientInfo

	buff, ok := params["clientinfo"]

	if !ok {
		return nil, errors.Wrap(ErrParams, "expect clientinfo param")
	}

	err = json.Unmarshal([]byte(buff), &ci)

	if err != nil {
		return nil, errors.Wrap(err, "unmarshal clientinfo param error")
	}

	return &wcTunnel{
		Self:     uuid.NewString(),
		Status:   Disconnected,
		Accounts: []string{account},
		SelfInfo: ci,
		URL:      u,
		Key:      key,
	}, nil
}

func fromContext(context []byte) (*wcTunnel, error) {
	var tunnel *wcTunnel
	err := json.Unmarshal(context, &tunnel)

	if err != nil {
		return nil, errors.Wrap(err, "unmarshal wcTunnel context error")
	}

	return tunnel, nil
}

func (tunnel *wcTunnel) send(topic string, data []byte) ([]byte, error) {

	var msg *socketMessage = nil

	if len(data) != 0 {
		encryptData, err := encrypt(data, tunnel.Key)

		if err != nil {
			return nil, err
		}

		payload, err := json.Marshal(encryptData)

		if err != nil {
			return nil, errors.Wrap(err, "marshal encryptionPayload error")
		}

		msg = &socketMessage{
			Topic:   topic,
			Type:    "pub",
			Payload: string(payload),
		}

	} else {
		msg = &socketMessage{
			Topic:   topic,
			Type:    "pub",
			Payload: "",
		}
	}

	buff, err := json.Marshal(msg)

	if err != nil {
		return nil, errors.Wrap(err, "marshal socketMessage error")
	}

	return buff, nil
}

func (tunnel *wcTunnel) Send(msg []byte, transport tun4go.Transport) error {

	if tunnel.Status != Connected {
		return errors.Wrap(ErrStatus, "send msg with invalid status %s", tunnel.Status)
	}

	buff, err := tunnel.send(tunnel.Peer, msg)

	if err != nil {
		return err
	}

	err = transport.Write(buff)

	if err != nil {
		return errors.Wrap(err, "write to transport error")
	}

	return nil
}

func (tunnel *wcTunnel) Recv(transport tun4go.Transport) ([]byte, error) {

Start:

	if tunnel.Status != Connected {
		return nil, errors.Wrap(ErrStatus, "send msg with invalid status %s", tunnel.Status)
	}

	data, err := transport.Read()

	if err != nil {
		return nil, errors.Wrap(err, "read from trasnport error")
	}

	buff, err := tunnel.read(data)

	if err != nil {
		return nil, errors.Wrap(err, "decode recv msg error : %s", string(data))
	}

	request, err := tunnel.readJSONRPCRequest(buff)

	if request.Method == "wc_sessionUpdate" {
		if err := tunnel.handleSessionUpdate(request); err != nil {
			return nil, err
		}
		goto Start
	}

	return buff, nil
}

func (tunnel *wcTunnel) handleSessionUpdate(request *jsonRPCRequest) error {

	if len(request.Params) != 1 {
		return errors.Wrap(ErrFormat, "wc_ssessionUpdate params number must be 1")
	}

	buff, err := json.Marshal(request.Params[0])

	if err != nil {
		return errors.Wrap(err, "marshal sessionUpdate request error")
	}

	var update *sessionUpdate

	err = json.Unmarshal(buff, &update)

	if err != nil {
		return errors.Wrap(err, "unmarshal sessionUpdate request error")
	}

	if update.Approved == false && tunnel.Status == Connected {
		tunnel.Status = Disconnected
		return errors.Wrap(ErrDisconnected, "peer %s disconnct", tunnel.Peer)
	}

	return nil
}

func (tunnel *wcTunnel) read(data []byte) ([]byte, error) {
	var msg *socketMessage

	err := json.Unmarshal(data, &msg)

	if err != nil {
		return nil, errors.Wrap(err, "unmarshal socketMessage error %s", string(data))
	}

	var encryptData *encryptionPayload

	err = json.Unmarshal([]byte(msg.Payload), &encryptData)

	if err != nil {
		return nil, errors.Wrap(err, "unmarshal encryptionPayload error %s", msg.Payload)
	}

	buff, err := encryptData.decrypt(tunnel.Key)

	return buff, err
}

func (tunnel *wcTunnel) Context() ([]byte, error) {
	buff, err := json.Marshal(&tunnel)

	if err != nil {
		return nil, errors.Wrap(err, "marshal wcTunnel error")
	}

	return buff, nil
}

// Disconnect send disconnect msg to peer
func (tunnel *wcTunnel) Disconnect(transport tun4go.Transport) error {
	return nil
}

func (tunnel *wcTunnel) subscribe(topic string, transport tun4go.Transport) error {
	msg := &socketMessage{
		Topic:   topic,
		Type:    "sub",
		Payload: "",
	}

	buff, err := json.Marshal(msg)

	if err != nil {
		return errors.Wrap(err, "marshal socketMessage error")
	}

	err = transport.Write(buff)

	if err != nil {
		return errors.Wrap(err, "write sub %s to transport error", topic)
	}

	return nil
}

func (tunnel *wcTunnel) Connect(transport tun4go.Transport) error {

	if tunnel.Status != Disconnected {
		return nil
	}

	tunnel.Status = Connecting

	err := tunnel.subscribe(tunnel.URL.Topic, transport)

	if err != nil {
		tunnel.Status = Disconnected
		return err
	}

	buff, err := transport.Read()

	if err != nil {
		tunnel.Status = Disconnected
		return errors.Wrap(err, "read sessionRequest error")
	}

	buff, err = tunnel.read(buff)

	if err != nil {
		tunnel.Status = Disconnected
		return err
	}

	request, err := tunnel.readJSONRPCRequest(buff)

	if err != nil {
		tunnel.Status = Disconnected
		return err
	}

	if request.Method != "wc_sessionRequest" {
		tunnel.Status = Disconnected
		return errors.Wrap(ErrMessage, "expect wc_sessionRequest but got %s", request.Method)
	}

	err = tunnel.handleSessionRequest(request, transport)

	if err != nil {
		tunnel.Status = Disconnected
		return err
	}

	tunnel.Status = Connected

	return nil
}

func (tunnel *wcTunnel) handleSessionRequest(request *jsonRPCRequest, transport tun4go.Transport) error {
	if len(request.Params) != 1 {
		return errors.Wrap(ErrFormat, "wc_sessionRequest params number must be 1")
	}

	buff, err := json.Marshal(request.Params[0])

	if err != nil {
		return errors.Wrap(err, "marshal sessionRequest request error")
	}

	var sr *sessionRequest

	err = json.Unmarshal(buff, &sr)

	if err != nil {
		return errors.Wrap(err, "unmarshal sessionRequest request error")
	}

	approver, ok := transport.(tun4go.Approver)

	approved := false

	if !ok {
		approved = true
	} else {
		approved = approver.Approve(buff)
	}

	if approved {
		tunnel.Peer = sr.PeerID
		tunnel.PeerInfo = sr.PeerMeta
	}

	if err := tunnel.approve(request.ID, approved, transport); err != nil {
		return err
	}

	if approved {
		return tunnel.subscribe(tunnel.Self, transport)
	}

	return nil
}

func (tunnel *wcTunnel) approve(id int64, approved bool, transport tun4go.Transport) error {

	rsp := &sessionResponse{
		PeerID:   tunnel.Self,
		PeerMeta: tunnel.SelfInfo,
		ChainID:  tunnel.ChainID,
		Approved: approved,
		Accounts: tunnel.Accounts,
	}

	rpc := &jsonRPCResponse{
		ID:      id,
		JSONRPC: "2.0",
		Result:  rsp,
	}

	buff, err := json.Marshal(rpc)

	if err != nil {
		return errors.Wrap(err, "marshal sessionResponse error")
	}

	buff, err = tunnel.send(tunnel.Peer, buff)

	if err != nil {
		return errors.Wrap(err, "encrypt sessionResponse error")
	}

	err = transport.Write(buff)

	if err != nil {
		return errors.Wrap(err, "write session response to transport error")
	}

	return nil
}

func (tunnel *wcTunnel) readJSONRPCRequest(buff []byte) (*jsonRPCRequest, error) {
	var request *jsonRPCRequest

	err := json.Unmarshal(buff, &request)

	if err != nil {
		return nil, errors.Wrap(err, "unmarshal handshake request error")
	}

	return request, nil
}
