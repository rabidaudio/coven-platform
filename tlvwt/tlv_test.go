package tlvwt

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	msg := TLVMessage{
		"foo": []byte{ 0xDE, 0xAD, 0xBE, 0xEF },
		"bar": []byte("BAR"),
		"baz": []byte{},
	}
	data, err := msg.EncodeToString()
	assert.NoError(t, err)
	// should be alphabetical
	assert.Equal(t, "bar", string(data[0:3]))
	assert.Equal(t, uint8(3), data[3])
	assert.Equal(t, "BAR", string(data[4:7]))

	assert.Equal(t, "baz", string(data[7:10]))
	assert.Equal(t, uint8(0), data[10])

	assert.Equal(t, "foo", string(data[11:14]))
	assert.Equal(t, uint8(4), data[14])
	assert.Equal(t, byte(0xDE), data[15])
	assert.Equal(t, byte(0xAD), data[16])
	assert.Equal(t, byte(0xBE), data[17])
	assert.Equal(t, byte(0xEF), data[18])

	decoded, err := DecodeString(data)
	assert.NoError(t, err)
	assert.EqualValues(t, msg, decoded)
}

func TestInvalid(t *testing.T) {
	a := TLVMessage{
		"toolong": []byte("asdf"),
	}
	_, err := a.EncodeToString()
	assert.ErrorIs(t, err, ErrInvalidKey{Key: "toolong"})

		b := TLVMessage{
		"s": []byte("asdf"),
	}
	_, err = b.EncodeToString()
	assert.ErrorIs(t, err, ErrInvalidKey{Key: "s"})

	data := make([]byte, 256)
	rand.Read(data)
	c := TLVMessage{
		"foo": data,
	}
	_, err = c.EncodeToString()
	assert.ErrorIs(t, err, ErrInvalidLength{Key: "foo"})
}

func TestEmpty(t *testing.T) {
	empty := TLVMessage{}
	data, err := empty.EncodeToString()
	assert.NoError(t, err)
	assert.Empty(t, data)
}
