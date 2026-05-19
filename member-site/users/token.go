package users

import (
	"crypto/ed25519"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/atlantacoven/coven-platform/tlvwt"
)

const ExpiresIn = 14 * 24 * time.Hour // 2 weeks

var privkey ed25519.PrivateKey

func init() {
	keydata, err := os.ReadFile("member-site/signing.pem")
	if err != nil {
		panic(fmt.Errorf("open private key: %w", err))
	}
	pk, err := x509.ParsePKCS8PrivateKey(keydata)
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
