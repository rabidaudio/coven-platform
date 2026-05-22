#ifndef ISO7816_4_H_
#define ISO7816_4_H_

#include <Arduino.h>

#include <SPI.h>
#include <PN532_SPI.h>
#include <PN532_SPI.cpp>
#include "PN532.h"

#ifndef NFC_CS_PIN
#define NFC_CS_PIN 15
#endif

PN532_SPI pn532spi(SPI, NFC_CS_PIN);
PN532 nfc(pn532spi);

#define STATUS_OK 0x9000

// Global buffer for storing messages across NFC
uint8_t _messagebuf[256];

struct MessageHeader {
    uint8_t cla;
    uint8_t ins;
    uint8_t p1;
    uint8_t p2;
};

#define MESSAGE_MAX_SIZE (255 - 5 - 1)

void beginNFC() {
  nfc.begin();
  size_t version = nfc.getFirmwareVersion();
  if (version == 0) {
    Serial.println(F("PN532 not found"));
    while (1) ;;
  }
  Serial.print(F("Firmware version: ")); Serial.println(version);
  nfc.SAMConfig();
}

bool cardInRange() {
    return nfc.inListPassiveTarget();
}

void debugPrintHex(uint8_t* ptr, size_t size) {
    for (size_t i = 0; i < size; i++) {
        if (ptr[i] < 16) Serial.print("0");
        Serial.print(ptr[i], HEX);
        if ((i+1) % 32 == 0) Serial.println();
    }
    Serial.println();
}

class Message {
    private:
    uint8_t dataSize;
    uint8_t resSize;

    void finalize() {
        if (dataSize == 0) {
            // exclude byte 4 Lc
            _messagebuf[4] = 0x00; // accept up to 256 bytes back
        } else {
            _messagebuf[4] = dataSize;
            _messagebuf[5 + dataSize] = 0x00; // accept up to 256 bytes back
        }
    }

    bool expand(uint8_t size) {
        if (size > MESSAGE_MAX_SIZE) return false; // too big
        if (((uint16_t)dataSize + size) > MESSAGE_MAX_SIZE) return false; // overflow
        dataSize += size;
        return true;
    }

    public:
    
    Message() {
        dataSize = 0;
        resSize = 0;
    }

    void setHeader(const MessageHeader* header) {
        _messagebuf[0] = header->cla;
        _messagebuf[1] = header->ins;
        _messagebuf[2] = header->p1;
        _messagebuf[3] = header->p2;
    }
    
    bool appendDataFrom(uint8_t* src, uint8_t size) {
        memcpy(_messagebuf + 5 + dataSize, src, size);
        return expand(size);
    }

    bool appendDataDirect(uint8_t** dataPtr, uint8_t sizeToBeAppended) {
        *dataPtr = _messagebuf + 5 + dataSize;
        return expand(sizeToBeAppended);
    }

    bool appendUInt16(uint16_t data) {
        uint8_t a = (uint8_t)(data >> 8);
        uint8_t b = (uint8_t)(data & 0xFF);
        _messagebuf[5 + dataSize] = a;
        _messagebuf[5 + dataSize + 1] = b;
        return expand(2);
    }

    bool send(uint8_t** responsePtr = NULL, uint8_t* responseSize = NULL) {
        finalize();
        resSize = 255; // max size we can handle in buffer
        uint8_t messageSize = dataSize + 4 + 1;
        if (dataSize != 0) messageSize += 1;
        Serial.print(F("send[")); Serial.print(messageSize); Serial.print(F("]:\t"));
        debugPrintHex(_messagebuf, messageSize);
        if (!nfc.inDataExchange(_messagebuf, messageSize, _messagebuf, &resSize)) {
            Serial.println(F("send failed"));
            return false;
        }
        Serial.print(F("recv[")); Serial.print(resSize); Serial.print(F("]:\t"));
        debugPrintHex(_messagebuf, resSize);
        if (responsePtr != NULL) {
            *responsePtr = _messagebuf;
        }
        if (responseSize != NULL) {
            *responseSize = resSize-2;
        }
        return true;
    }

    uint16_t getResponseStatus() {
        uint8_t* statusIdx = _messagebuf + resSize - 2;
        return ((uint16_t)(statusIdx[0]) << 8) + ((uint16_t) (statusIdx[1]));
    }

    bool isResponseOk() {
        return getResponseStatus() == STATUS_OK;
    }
};

#endif // define ISO7816_4_H_
