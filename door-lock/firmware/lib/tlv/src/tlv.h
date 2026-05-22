#ifndef TLV_H_
#define TLV_H_

#include <Arduino.h>

#define TLV_TAG_LEN 3
#define TLV_SIZE_LEN 1
#define TLV_HEADER_LEN (TLV_TAG_LEN + TLV_SIZE_LEN)

/*
Implements the TLV encoding/decoding mechanism with 3-byte tags
and 1-byte lengths. Operates entirely in-place on top of the
original buffer. The same tag cannot be written multiple times.
*/
class TLVMessage {
private:
  uint8_t* _buf;
  size_t _cap;
  size_t _len;

public:
  TLVMessage(uint8_t* buf, size_t len, size_t cap) {
    _buf = buf;
    _len = len;
    _cap = cap;
  }

  size_t getLength() { return _len; }

  bool getTag(const char* tag, uint8_t** out, uint8_t* outlen) {
    uint8_t* loc = findTag(tag);
    if (loc == NULL) {
      return false;
    }
    if (out != NULL)
      *out = loc + TLV_HEADER_LEN;
    if (outlen != NULL)
      *outlen = loc[TLV_TAG_LEN];
    return true;
  }

  bool appendFrom(const char* tag, uint8_t* data, uint8_t datalen) {
    size_t toWrite = 4 + datalen;
    if (_len + toWrite > _cap) {
      return false; // no more room
    }
    memcpy(_buf + _len, tag, TLV_TAG_LEN);
    _buf[_len + TLV_TAG_LEN] = datalen;
    memcpy(_buf + _len + TLV_HEADER_LEN, data, datalen);
    _len += (TLV_HEADER_LEN + datalen);
    return true;
  }

  bool appendDataDirect(const char* tag, uint8_t** dataPtr,
                        uint8_t sizeToBeAppended) {
    size_t toWrite = TLV_HEADER_LEN + sizeToBeAppended;
    if (_len + toWrite > _cap) {
      return false; // no more room
    }
    memcpy(_buf + _len, tag, TLV_TAG_LEN);
    _buf[_len + TLV_TAG_LEN] = sizeToBeAppended;
    *dataPtr = _buf + _len + TLV_HEADER_LEN;
    _len += (TLV_HEADER_LEN + sizeToBeAppended);
    return true;
  }

private:
  // returns the pointer to the start of tag or NULL if not found
  uint8_t* findTag(const char* tag) {
    uint8_t* pos = _buf;
    size_t rem = _len;
    while (true) {
      if (rem < TLV_HEADER_LEN) {
        return NULL; // not found
      } else if (memcmp(tag, pos, TLV_TAG_LEN) == 0) {
        return pos; // found
      } else {
        // try next
        uint8_t size = pos[TLV_TAG_LEN];
        rem -= (TLV_HEADER_LEN + size);
        pos += (TLV_HEADER_LEN + size);
      }
    }
  }
};

#endif // TLV_H_
