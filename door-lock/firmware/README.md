# Door Lock

This project implements an electronic NFC door lock for the makerspace from scratch,
allowing active members access to the space via their phone.

The microprocessor is an ESP32. It does connect to WIFI, but only to periodically synchronize
the internal RTC via NTP. All authentication is done offline.

## Handshake algorithm

The security requirements for the door lock mechanism:
- An attacker should not be able to generate a valid key in a reasonable time
- The NFC communication should be considered cleartext. An attacker who eavesdropped on a successful NFC handshake should not be able to reuse a previously valid key
- An attacker who spoofs the lock controller should not be able to steal a valid key from the app
- The lock controller should be able to verify authenticity offline, to avoid a network-based attack

Based on these requirements, the following algorithm is used:
1. The app authenticates via HTTPS with the server using a typical mechanism (password, oauth, etc).
2. The server returns a `Token`, which is info about the user, signed by the server using ED25519. Similar in design to a JWT token but in binary rather than JSON due to embedded performance.
3. The app saves this `Token` securely on the device. It periodically re-authenticates to avoid expiration.
4. The user presents the phone to the door controller. The phone and door controller communicate via standards described in ISO 14443-4 and ISO 7816-4.
5. The door sends an `AID` identifying it as implementing this (proprietary) algorithm. The app has associated itself with this `AID` so that it executes in response to the message.
6. The door sends an authentication `Challenge` to the app, which is it's a random nonce and a public key to use for secure transfer of the `Token`, both signed with ED25519 to verify it as an authentic challenge.
7. The app verifies the signature, then encrypts the `Token` and sends it. Secure transfer of the `UserSecret` is done using HPKE as defined in [RFC 9180](https://datatracker.ietf.org/doc/rfc9180/), using `DHKEM(X25519, HKDF-SHA256)/HKDF-SHA256/AES-128-GCM`.
8. The door decrypts the message. Then it verifies that the nonce is correct, that the `Token` is signed by the server, and that it has not expired.
9.  If all of those checks pass, it unlocks the door for a brief number of seconds. It sends a message to the app indicating if the door was unlocked.

An example of this algorithm can be found in `keygen/algorithm_test.go`.

## Useful Reference

- https://learn.adafruit.com/adafruit-pn532-rfid-nfc/about-nfc
- https://developer.android.com/develop/connectivity/nfc/hce
- http://www.emutag.com/iso/14443-3.pdf
- http://www.emutag.com/iso/14443-4.pdf
- https://www.freecalypso.org/pub/GSM/ISO7816/ISO_7816-4_2005.pdf
- https://cdn-shop.adafruit.com/datasheets/pn532um.pdf
- https://datatracker.ietf.org/doc/rfc9180/
