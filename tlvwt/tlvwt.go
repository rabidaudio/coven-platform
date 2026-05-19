// Package twlvwt implements a JSON Web Token (JWT)-like token interface,
// but using TLV as the encoding mechanism in place of JSON, to provide
// better support for embedded platforms. Unlike JWT, the message is not
// base64 encoded but rather sent in it's raw binary format.
//
// The specific variant of TLV has 3 byte tags (which correspond to
// standard JWT claims such as "alg" or "jti"), a 1-byte unsigned length
// (making the maximum field size 255 bytes), and a payload. Data is
// in Big-Endian (network order). Numeric dates should be in 64-bit unsigned
// seconds since the Unix epoch. Strings should not include null terminators.
// Tags _should_ be in alphabetical order, except for the trailing "sig" tag.
//
// The signature of the message is computed from the encoded binary message
// excluding the "sig" tag. The signature is appended as a TLV field at the end
// of the message under the "sig" tag.
//
// https://www.rfc-editor.org/rfc/rfc7519
// https://en.wikipedia.org/wiki/Type%E2%80%93length%E2%80%93value
// https://en.wikipedia.org/wiki/JSON_Web_Token
package tlvwt

import (
	"bytes"
	"io"
	"time"
)

type Token TLVMessage

func (t Token) SetIssuer(iss string) {
	TLVMessage(t).SetString("iss", iss)
}

func (t Token) SetSubject(sub string) {
	TLVMessage(t).SetString("sub", sub)
}

func (t Token) SetAudience(aud string) {
	TLVMessage(t).SetString("aud", aud)
}

func (t Token) SetExpiration(exp time.Time) {
	TLVMessage(t).SetDate("exp", exp)
}

func (t Token) SetNotBefore(nbf time.Time) {
	TLVMessage(t).SetDate("nbf", nbf)
}

func (t Token) SetIssuedAt(iat time.Time) {
	TLVMessage(t).SetDate("iat", iat)
}

func (t Token) SetID(id string) {
	TLVMessage(t).SetString("jti", id)
}

func (t Token) SetType(typ string) {
	if typ == "" {
		typ = "TWT"
	}
	TLVMessage(t).SetString("typ", typ)
}

func (t Token) SetAlgorithm(alg string) {
	TLVMessage(t).SetString("alg", alg)
}

// --

func (t Token) SignAndEncode(w io.Writer, sign SigningAlgorithm) (int, error) {
	if sign == nil {
		sign = AlgoNone{}
	}
	sign.Tag(t)
	msg, err := TLVMessage(t).EncodeToString()
	if err != nil {
		return 0, err
	}
	sig, err := sign.Sign(msg)
	if err != nil {
		return 0, err
	}
	var sigmsg []byte
	if len(sig) > 0 {
		sigmsg, err = TLVMessage{"sig": sig}.EncodeToString()
		if err != nil {
			return 0, err
		}
	}

	n := 0
	nn, err := w.Write(msg)
	n += nn
	if err != nil {
		return n, err
	}
	if len(sig) > 0 {
		nn, err = w.Write(sigmsg)
		n += nn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (t Token) SignAndEncodeString(sign SigningAlgorithm) ([]byte, error) {
		buf := bytes.Buffer{}
	_, err := t.SignAndEncode(&buf, sign)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t Token) String() string {
	return TLVMessage(t).String()
}
