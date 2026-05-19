package tlvwt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"
)

// ErrInvalidKey is returned when encoding a TLV object with a tag
// whose length is not 3 bytes.
type ErrInvalidKey struct { Key string }
func (err ErrInvalidKey) Error() string {
	return fmt.Errorf("invalid key in TLV: %v", err.Key).Error()
}

// ErrInvalidKey is returned when encoding a TLV object with a field
// value whose length is greater than 255 bytes.
type ErrInvalidLength struct { Key string }
func (err ErrInvalidLength) Error() string {
	return fmt.Errorf("invalid value for key in TLV: %v", err.Key).Error()
}

type TLVMessage map[string][]byte

func (t TLVMessage) SetString(key, val string) {
	t[key] = []byte(val)
}

func (t TLVMessage) GetString(key string) string {
	return string(t[key])
}

func (t TLVMessage) SetUInt64(key string, val uint64) {
	t[key] = binary.BigEndian.AppendUint64([]byte{}, val)
}

func (t TLVMessage) GetUInt64(key string) uint64 {
	return binary.BigEndian.Uint64(t[key])
}

func (t TLVMessage) SetDate(key string, tt time.Time) {
	t.SetUInt64(key, uint64(tt.Unix()))
}

func (t TLVMessage) GetDate(key string) time.Time {
	return time.Unix(int64(t.GetUInt64(key)), 0)
}


func (msg TLVMessage) Encode(w io.Writer) (int, error) {
	keys := slices.Collect(maps.Keys(msg))
	slices.Sort(keys)
	// validate before writing anything
	for _, k := range keys {
		if len(k) != 3 {
			return 0, ErrInvalidKey{Key: k}
		}
		if len(msg[k]) >= 256 {
			return 0, ErrInvalidLength{Key: k}
		}
	}

	n := 0
	for _, k := range keys {
		nn, err := w.Write([]byte(k))
		n += nn
		if err != nil {
			return n, err
		}
		nn, err = w.Write([]byte{ uint8(len(msg[k])) })
		n += nn
		if err != nil {
			return n, err
		}
		nn, err = w.Write(msg[k])
		n += nn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (msg TLVMessage) EncodeToString() ([]byte, error) {
	buf := bytes.Buffer{}
	_, err := msg.Encode(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeString(data []byte) (TLVMessage, error) {
	msg := TLVMessage{}
	for {
		if len(data) == 0 {
			return msg, nil
		}
		if len(data) < 4 {
			return nil, ErrInvalidLength{}
		}
		tag := string(data[0:3])
		size := int(data[3])
		if (len(data) - 4) < size {
			return nil, ErrInvalidLength{}
		}
		msg[tag] = data[4:4+size]
		data = data[4+size:]
	}
}

func (msg TLVMessage) String() string {
	keys := slices.Collect(maps.Keys(msg))
	slices.Sort(keys)
	buf := strings.Builder{}
	for i, k := range keys {
		buf.Write([]byte(k))
		fmt.Fprintf(&buf, "[%v]: ", len(msg[k]))
		if isASCII(msg[k]) {
			buf.Write([]byte(msg[k]))
		} else {
			buf.Write([]byte("0x"))
			buf.Write([]byte(hex.EncodeToString(msg[k])))
		}
		if i+1 != len(keys) {
			buf.Write([]byte(", "))
		}
	}
	return buf.String()
}

func isASCII(data []byte) bool {
	for _, c := range data {
		if c < 9 || (c > 13 && c < 32) || c > 126 {
			return false
		}
	}
	return true
}
