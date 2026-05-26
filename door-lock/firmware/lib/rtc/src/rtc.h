#ifndef RTC_H_
#define RTC_H_

#include <Arduino.h>

#include <NTP3.h>
#if defined ARDUINO_ESP32_THING
#include <WiFi.h>
#endif

#include "_keys.h"

WiFiUDP wifiUdp;
NTP3 ntp(wifiUdp);

class RTC {
public:
  void begin() {
    WiFi.begin(WIFI_SSID, WIFI_PASS);

    ntp.updateInterval(100 * 60 * 60); // 1 hour NTP sync
    ntp.syncRTC(true);                 // update ESP32 clock
    ntp.begin();
  }

  void tryConnect() {
    size_t attempts = 10;
    while (attempts > 0) {
      Serial.println("Connecting ...");
      delay(500);
      if (isConnected()) {
        Serial.println("Connected");
        return;
      }
      attempts--;
    }
  }

  bool isConnected() { return WiFi.status() == WL_CONNECTED; }

  void update() {
    if (!isConnected()) {
      tryConnect();
    }
    if (!ntp.update()) {
      Serial.println("warning: unable to update time. Trusting RTC");
    }
  }

  uint64_t getTime() { return (uint64_t)ntp.epoch(); }
};

#endif // RTC_H_
