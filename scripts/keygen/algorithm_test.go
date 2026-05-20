package main

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hpke"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"testing"
	"time"

	"github.com/atlantacoven/coven-platform/tlvwt"
	"github.com/stretchr/testify/assert"
)

const NonceSize = 8

var KEM = hpke.DHKEM(ecdh.X25519())
var KDF = hpke.HKDFSHA256()
var AEAD = hpke.AES128GCM()

var ExpiresAfter = time.Duration(14 * 24 * time.Hour)

type Server struct {
	PublicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

func NewServer() *Server {
	s := Server{}
	var err error
	s.PublicKey, s.privateKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return &s
}

func (s *Server) GenerateUserToken(uid int) []byte {
	expires := time.Now().Add(ExpiresAfter)

	token := tlvwt.Token{}
	token.SetType("TWT")
	token.SetIssuer(DOMAIN)
	token.SetExpiration(expires)
	tlvwt.TLVMessage(token).SetUInt64("uid", uint64(uid))
	msg, err := token.SignAndEncodeString(tlvwt.Ed25519Algorithm{Key: s.privateKey})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Token: %v\n", token)
	fmt.Printf("TokenData[%v]: %x\n", len(msg), msg)
	return msg
}

type DoorLock struct {
	PublicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey

	recvKey hpke.PrivateKey

	nonce []byte
}

func NewDoorLock() *DoorLock {
	var d DoorLock
	var err error
	d.PublicKey, d.privateKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return &d
}

func (dl *DoorLock) GenerateChallenge() []byte {
	dl.nonce = GenerateNonce()

	var err error
	dl.recvKey, err = KEM.GenerateKey()
	if err != nil {
		panic(err)
	}

	challenge := tlvwt.Token{
		"nce": dl.nonce,
		"pub": dl.recvKey.PublicKey().Bytes(),
	}
	data, err := challenge.SignAndEncodeString(tlvwt.Ed25519Algorithm{Key: dl.privateKey})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Challenge: %v\n", challenge)
	fmt.Printf("ChallengeData[%v]: %x\n", len(data), data)
	return data
}

func (dl *DoorLock) DecodeEncryptedToken(data []byte) []byte {
	enc := data[0:32]
	ct := data[32:]
	r, err := hpke.NewRecipient(enc, dl.recvKey, KDF, AEAD, []byte(INFO))
	if err != nil {
		panic(err)
	}
	dec, err := r.Open(nil, ct)
	if err != nil {
		panic(err)
	}
	nonce := dec[0:NonceSize]
	fmt.Printf("Nonce[%v]: %x\n", len(nonce), nonce)
	if subtle.ConstantTimeCompare(nonce, dl.nonce) != 1 {
		panic("invalid nonce")
	}
	decryptedSize := int(dec[NonceSize])
	fmt.Printf("TokenSize: %x\n", decryptedSize)
	token := dec[NonceSize+1:]
	fmt.Printf("TokenData[%v]: %x\n", len(token), token)
	return token
}

func (dl *DoorLock) VerifyToken(token []byte, serverpub ed25519.PublicKey) int {
	tmsg, err := tlvwt.VerifyAndExtractToken(token, func(msg, sig []byte) bool {
		return ed25519.Verify(serverpub, msg, sig)
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Token: %v\n", tmsg)
	if tmsg.GetString("typ") != "TWT" {
		panic("invalid type")
	}
	if tmsg.GetString("iss") != DOMAIN {
		panic("invalid issuer")
	}
	expires := tmsg.GetDate("exp")
	if expires.Compare(time.Now()) < 0 {
		panic("expired")
	}
	uid := int(tmsg.GetUInt64("uid"))
	return uid
}

type App struct {
	token []byte
}

func (a *App) EncryptToken(challenge, doorpub []byte) []byte {
	ch, err := tlvwt.VerifyAndExtractToken(challenge, func(msg, sig []byte) bool {
		return ed25519.Verify(doorpub, msg, sig)
	})
	if err != nil {
		panic(err)
	}
	nonce := ch["nce"]
	fmt.Printf("Nonce[%v]: %x\n", len(nonce), nonce)
	pkr, err := KEM.NewPublicKey(ch["pub"])
	if err != nil {
		panic(err)
	}
	enc, s, err := hpke.NewSender(pkr, KDF, AEAD, []byte(INFO))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Enc[%v]: %x\n", len(enc), enc)
	var msg []byte
	msg = append(msg, nonce...)
	msg = append(msg, byte(len(a.token)))
	msg = append(msg, a.token...)
	ct, err := s.Seal(nil, msg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("CipherText[%v]: %x\n", len(ct), ct)
	var e []byte
	e = append(e, enc...)
	e = append(e, ct...)
	return e
}

func TestAlgorithm(t *testing.T) {
	userId := 1234

	server := NewServer()
	door := NewDoorLock()
	app := App{}

	// App authenticates with server and receives token. It saves this securely to
	// the device
	app.token = server.GenerateUserToken(userId)

	// When at the door, door issues challenge and app responds with the token, encrypted via
	// HPKE
	challenge := door.GenerateChallenge()
	encryptedToken := app.EncryptToken(challenge, door.PublicKey)

	// The door decrypts and verifies this token, then unlocks the door
	token := door.DecodeEncryptedToken(encryptedToken)
	uid := door.VerifyToken(token, server.PublicKey)

	assert.Equal(t, uid, userId)
}

func GenerateNonce() []byte {
	b := make([]byte, NonceSize)
	rand.Read(b)
	return b
}
