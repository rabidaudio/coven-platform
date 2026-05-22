#ifndef RTC_H_
#define RTC_H_

#include <Arduino.h>

#include <NTP3.h>
#if defined ARDUINO_ESP32_THING
#include <WiFi.h>
#endif

WiFiUDP wifiUdp;
NTP3 ntp(wifiUdp);

class RTC {
public:
  void begin() {
    ntp.updateInterval(100 * 60 * 60); // 1 hour NTP sync
    ntp.syncRTC(true);                 // update ESP32 clock
    ntp.begin();
  }

  void update() {
    if (!ntp.update()) {
      Serial.println("warning: unable to update time. Trusting RTC");
    }
  }

  uint64_t getTime() { return (uint64_t)ntp.epoch(); }
};

#endif // RTC_H_
