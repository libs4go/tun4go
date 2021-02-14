package wc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/libs4go/errors"
)

type socketMessage struct {
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type encryptionPayload struct {
	Data string `json:"data"`
	Hmac string `json:"hmac"`
	IV   string `json:"iv"`
}

type jsonRPCRequest struct {
	ID      int64         `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type jsonRPCResponse struct {
	ID      int64         `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type sessionRequest struct {
	PeerID   string      `json:"peerId"`
	PeerMeta *clientInfo `json:"peerMeta"`
	ChainID  *int64      `json:"chainId"`
}

type sessionResponse struct {
	PeerID   string      `json:"peerId"`
	PeerMeta *clientInfo `json:"peerMeta"`
	ChainID  int64       `json:"chainId"`
	Approved bool        `json:"approved"`
	Accounts []string    `json:"accounts"`
}

type sessionUpdate struct {
	ChainID  int64    `json:"chainId"`
	Approved bool     `json:"approved"`
	Accounts []string `json:"accounts"`
}

func (payload *encryptionPayload) decrypt(key []byte) ([]byte, error) {

	data, err := hex.DecodeString(payload.Data)

	if err != nil {
		return nil, errors.Wrap(err, "decode data %s error", payload.Data)
	}

	hmac, err := hex.DecodeString(payload.Hmac)

	if err != nil {
		return nil, errors.Wrap(err, "decode hmac %s error", payload.Hmac)
	}

	iv, err := hex.DecodeString(payload.IV)

	if err != nil {
		return nil, errors.Wrap(err, "decode iv %s error", payload.IV)
	}

	expect := computeHmac(data, iv, key)

	if !bytes.Equal(expect, hmac) {
		return nil, errors.Wrap(ErrHMAC, "hmac expect %s got %s", hex.EncodeToString(expect), payload.Hmac)
	}

	block, _ := aes.NewCipher(key)

	decrypter := cipher.NewCBCDecrypter(block, iv[:])

	decrypter.CryptBlocks(data, data)

	return pkcs5Trimming(data), nil
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func pkcs5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func encrypt(data []byte, key []byte) (*encryptionPayload, error) {
	var iv [16]byte
	_, err := rand.Read(iv[:])

	if err != nil {
		return nil, errors.Wrap(err, "generate iv error")
	}

	plainData := pkcs5Padding(data, aes.BlockSize, len(data))

	block, _ := aes.NewCipher(key)

	cipherData := make([]byte, len(plainData))

	encrypter := cipher.NewCBCEncrypter(block, iv[:])

	encrypter.CryptBlocks(cipherData, plainData)

	mac := hex.EncodeToString(computeHmac(cipherData, iv[:], key))

	return &encryptionPayload{
		Data: hex.EncodeToString(cipherData),
		Hmac: mac,
		IV:   hex.EncodeToString(iv[:]),
	}, nil
}

func computeHmac(payload []byte, iv []byte, key []byte) []byte {
	data := append(payload, iv...)

	mac := hmac.New(sha256.New, key)

	mac.Write(data)

	return mac.Sum(nil)
}
