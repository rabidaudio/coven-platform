#include <Arduino.h>

#if defined ARDUINO_ESP32_THING
#define NFC_CS_PIN 2
#define BUZZER_PIN 13
#define LOCK_PIN 12
#endif

#include "_keys.h"
#include "iso7816_4.h"
#include "key_verification.h"
#include "rtc.h"
#include "tlv.h"

RTC rtc;
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
void failBeep();
void unlockDoor();

void setup() {
  Serial.begin(115200);

  pinMode(BUZZER_PIN, OUTPUT);
  pinMode(LOCK_PIN, OUTPUT);
  digitalWrite(BUZZER_PIN, LOW);
  digitalWrite(LOCK_PIN, LOW);

  WiFi.begin(WIFI_SSID, WIFI_PASS);
  while (WiFi.status() != WL_CONNECTED) {
    Serial.println("Connecting ...");
    delay(500);
  }
  Serial.println("Connected");

  rtc.begin();
  verifier.begin(&rtc);
  beginNFC();
}

void loop() {
  rtc.update();

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
  } else {
    Serial.print(F("verification failed: "));
    Serial.println(res, HEX);
    writeDoorLockStatus(&doorStatusMsg, 0x6600 | (uint8_t)res);
  }
  // don't care if it goes through or not
  doorStatusMsg.send();

  if (res == 0) {
    unlockDoor();
  } else {
    failBeep();
  }
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

void failBeep() {
  for (size_t i = 0; i < 3; i++) {
    digitalWrite(BUZZER_PIN, HIGH);
    delay(100);
    digitalWrite(BUZZER_PIN, LOW);
    delay(100);
  }
  delay(1000);
}

void unlockDoor() {
  digitalWrite(LOCK_PIN, HIGH);
  delay(100);
  digitalWrite(BUZZER_PIN, HIGH);
  delay(1000);
  digitalWrite(BUZZER_PIN, LOW);
  delay(5000);
  digitalWrite(LOCK_PIN, LOW);
}
