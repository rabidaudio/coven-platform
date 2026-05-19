package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	"github.com/atlantacoven/coven-platform/member-site/users"
	"github.com/atlantacoven/coven-platform/tlvwt"
	"github.com/stretchr/testify/assert"
)

func TestPostSession(t *testing.T) {
	s := TestServer(t)
	u1 := users.Fixture(t, s)

	// test find
	res, err := s.Request("POST", "/session", `{"email":"julien@example.com","password":"password"}`)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, u1.Id, int(res.Data["id"].(float64)))
	if token64, ok := res.Data["token"].(string); ok {
		tokendata, err := base64.StdEncoding.DecodeString(token64)
		assert.NoError(t, err)
		token, err := tlvwt.DecodeString([]byte(tokendata))
		assert.NoError(t, err)
		assert.Equal(t, "TWT", string(token["typ"]))
		assert.Equal(t, "thecoven.space", string(token["iss"]))
		assert.GreaterOrEqual(t, token.GetDate("exp"), time.Now().Add(users.ExpiresIn).Add(-1*time.Second))
		assert.Equal(t, uint64(u1.Id), token.GetUInt64("uid"))
		assert.Equal(t, ed25519.SignatureSize, len(token["sig"]))
	} else {
		t.Fatal("token was not a string")
	}
}
