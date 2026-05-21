#include <Arduino.h>

#if defined ARDUINO_ESP32_THING
#define NFC_CS_PIN 2
#elif defined ARDUINO_TEENSY31
#define NFC_CS_PIN 15
#endif

#include "NTP3.h"
#include <WiFi.h>

WiFiUDP wifiUdp;
NTP3 ntp(wifiUdp);

void setup() {
  Serial.begin(115200);
  WiFi.begin(ssid, pass);
  while (WiFi.status() != WL_CONNECTED) {
    Serial.println("Connecting ...");
    delay(500);
  }
  Serial.println("Connected");
  ntp.updateInterval(300); // TODO: less frequent
  ntp.syncRTC(true);
  ntp.begin();
  Serial.println("start NTP");
  Serial.print("time_t size: ");
  Serial.println(sizeof(time_t));
}

void loop() {
  if (ntp.update()) {
    Serial.println(ntp.formattedTime("%d. %B %Y")); // dd. Mmm yyyy
    Serial.println(ntp.formattedTime("%A %T"));     // Www hh:mm:ss
    Serial.print("epoch: ");
    Serial.println(ntp.epoch());
  } else {
    Serial.println("No valid time data");
  }
  delay(1000);
}

/*
#include "iso7816_4.h"
#include "key_verification.h"
#include "tlv.h"

KeyVerification verifier;

const MessageHeader SELECT_AID_MESSAGE = {
    0x00, // CLA
    0xA4, // INS: SELECT command
    0x04, // p1: the command data contains a DF name (the AID)
    0x00, // p2
};

const MessageHeader GENERAL_AUTH_MESSAGE = {
    0x00, // CLA
    0x86, // INS: GENERAL AUTHENTICATE
    0x00,
    0x00, // params (these are key and algorithm ids according to spec, but have
          // user-defined meanings)
};

const MessageHeader DOOR_STATUS_MESSAGE = {
    0xFA, // proprietary cla value (bit8=1)
    0x01, // proprietary inc
};

int writeGeneralAuthenticate(Message* msg);
bool writeDoorLockStatus(Message* msg, uint16_t status);

void setup() {
  Serial.begin(115200);
  verifier.begin();
  beginNFC();
}

void loop() {
  // wait for a card in range
  if (!cardInRange()) {
    delay(100);
    return;
  }

  // send SELECT message with AID
  Message selectMsg;
  selectMsg.setHeader(&SELECT_AID_MESSAGE);
  selectMsg.appendDataFrom((uint8_t*)AID, AID_LEN);

  uint8_t recvLen;
  uint8_t* responseBuf;
  if (!selectMsg.send(&responseBuf, &recvLen)) {
    return; // send failed
  }
  if (!selectMsg.isResponseOk()) {
    Serial.println(F("not ok"));
    return;
  }

  // Send GENERAL AUTHENTICATE message with Challenge and PubKey
  Message authMsg;
  if (writeGeneralAuthenticate(&authMsg) != 0) {
    Serial.println(F("build message failed"));
    return;
  }
  if (!authMsg.send(&responseBuf, &recvLen)) {
    return; // send failed
  }
  if (!authMsg.isResponseOk()) {
    Serial.println(F("not ok"));
    return;
  }

  // Check the result
  Serial.println("enc:");
  debugPrintHex(responseBuf, PUB_KEY_SIZE);
  Serial.println("ct:");
  debugPrintHex(responseBuf + PUB_KEY_SIZE, recvLen - PUB_KEY_SIZE);

  int res = verifier.decryptAndVerifyToken(responseBuf, recvLen);
  Message doorStatusMsg;
  if (res == 0) {
    Serial.println(F("verification success"));
    writeDoorLockStatus(&doorStatusMsg, STATUS_OK);
    // TODO: unlock door
  } else {
    Serial.print(F("verification failed: "));
    Serial.println(res);
    writeDoorLockStatus(&doorStatusMsg, 0x6600 | (uint16_t)res);
  }
  // don't care if it goes through or not
  doorStatusMsg.send();

  delay(1000);
}

int writeGeneralAuthenticate(Message* msg) {
  int res;
  uint8_t* outbuf;
  msg->setHeader(&GENERAL_AUTH_MESSAGE);
  if (!msg->appendDataDirect(&outbuf, CHALLENGE_SIZE))
    return -1;
  res = verifier.generateChallenge(outbuf);
  if (res != 0)
    return res;
  return 0;
}

bool writeDoorLockStatus(Message* msg, uint16_t status) {
  msg->setHeader(&DOOR_STATUS_MESSAGE);
  return msg->appendUInt16(status);
}
*/
