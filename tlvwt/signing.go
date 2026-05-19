package tlvwt

import (
	"crypto/ed25519"
	"crypto/sha256"
	"hash"
)

type SigningAlgorithm interface {
	Tag(t Token)
	Sign(msg []byte) ([]byte, error)
}

type AlgoNone struct{}

func (AlgoNone) Tag(t Token) {
	t.SetAlgorithm("none")
}

func (AlgoNone) Sign(msg []byte) ([]byte, error) {
	return []byte{}, nil
}

type algoHMAC struct {
	sha  func() hash.Hash
	name string
}

func (a *algoHMAC) Tag(t Token) {
	t.SetAlgorithm(a.name)
}

func (a *algoHMAC) Sign(msg []byte) ([]byte, error) {
	h := a.sha()
	return h.Sum(msg), nil
}

func HS256() SigningAlgorithm {
	return &algoHMAC{sha: sha256.New, name: "HS256"}
}

// recently specified: https://datatracker.ietf.org/doc/html/rfc9864#section-4.1.2
type Ed25519Algorithm struct {
	Key ed25519.PrivateKey
}

func (Ed25519Algorithm) Tag(t Token) {
	t.SetAlgorithm("Ed25519")
}

func (a Ed25519Algorithm) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(a.Key, msg), nil
}
