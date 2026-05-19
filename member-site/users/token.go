package users

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/atlantacoven/coven-platform/tlvwt"
	"github.com/integralist/go-findroot/find"
)

const ExpiresIn = 14 * 24 * time.Hour // 2 weeks

var privkey ed25519.PrivateKey

func init() {
	root, err := find.Repo()
	if err != nil {
		panic(fmt.Errorf("find project root: %w", err))
	}
	path := root.Path + "/member-site/signing.pem"
	keydata, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("open private key: %w", err))
	}
	block, _ := pem.Decode(keydata)
	if block == nil || block.Type != "ED25519 PRIVATE KEY" {
		panic(fmt.Errorf("invalid private key"))
	}
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("decode private key: %w", err))
	}
	privkey = pk.(ed25519.PrivateKey)
}

// GenerateToken generates a TLVWT token to be used to unlock the door
func (u *User) GenerateToken() ([]byte, error) {
	token := tlvwt.Token{}
	token.SetType("TWT")
	token.SetIssuer("thecoven.space")
	token.SetExpiration(time.Now().Add(ExpiresIn))
	tlvwt.TLVMessage(token).SetUInt64("uid", uint64(u.Id))
	return token.SignAndEncodeString(tlvwt.Ed25519Algorithm{Key: privkey})
}
