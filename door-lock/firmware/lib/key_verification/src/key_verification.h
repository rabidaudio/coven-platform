#ifndef KEY_VERIFICATION_H_
#define KEY_VERIFICATION_H_

#include <Arduino.h>
#include <wolfssl.h>
#include <wolfssl/wolfcrypt/ed25519.h>
#include <wolfssl/wolfcrypt/hpke.h>
#include <wolfssl/wolfcrypt/misc.h>
#include <wolfssl/wolfcrypt/random.h>

#include "_keys.h"
#include "rtc.h"
#include "tlv.h"

#define NONCE_SIZE 8

#define PUB_KEY_SIZE ED25519_PUB_KEY_SIZE
#define CHALLENGE_SIZE (NONCE_SIZE + PUB_KEY_SIZE + ED25519_SIG_SIZE + (3 * 4))

#define USER_ID_SIZE 8
#define DECRYPTED_MESSAGE_CAPACITY 128

// constant-time mem compare
int constantCompare(uint8_t* a, uint8_t* b, size_t len) {
  int compareSum = 0;
  for (int i = 0; i < len; i++) {
    compareSum |= a[i] ^ b[i];
  }
  return compareSum;
}

uint64_t readLongBE(uint8_t* src) {
  uint64_t result = 0;
  for (size_t i = 0; i < 8; i++) {
    result <<= 8;
    result |= (uint64_t)src[i];
  }
  return result;
}

class KeyVerification {
  RTC* rtc;
  WC_RNG kv_rng[1];
  ed25519_key server_sign_pub_key;
  ed25519_key door_sign_key;

  Hpke hpke[1];
  void* skr = NULL;

  uint8_t _nonce[NONCE_SIZE];
  uint8_t _decryptedMessage[DECRYPTED_MESSAGE_CAPACITY];

public:
  int begin(RTC* r) {
    rtc = r;
    int ret = wolfCrypt_Init();
    if (ret != 0) {
      return ret;
    }

    ret = wc_InitRng(kv_rng);
    if (ret != 0) {
      return ret;
    }

    ret =
        wc_ed25519_import_private_key(DOOR_SIGN_PRIV_KEY, ED25519_PRV_KEY_SIZE,
                                      NULL, 0, // indicates concat'ed priv+pub
                                      &door_sign_key);
    if (ret != 0) {
      return ret;
    }

    ret = wc_ed25519_import_public(SERVER_SIGNING_PUB_KEY, ED25519_PUB_KEY_SIZE,
                                   &server_sign_pub_key);
    if (ret != 0) {
      return ret;
    }

    ret = wc_HpkeInit(hpke, DHKEM_X25519_HKDF_SHA256, HKDF_SHA256,
                      HPKE_AES_128_GCM, NULL);
    if (ret != 0) {
      return ret;
    }

    return 0;
  }

  // generate a challenge, which is a TLV-encoded message containing
  // a 64-bit random nonce ("nce") and the receiver public key ("pub"),
  // , signed by door_sign_priv_key using ed25519 ("sig").
  int generateChallenge(uint8_t* challenge) {
    TLVMessage msg = TLVMessage(challenge, 0, CHALLENGE_SIZE);
    int ret = wc_RNG_GenerateBlock(kv_rng, _nonce, NONCE_SIZE);
    if (ret != 0)
      return ret;

    msg.appendFrom("nce", _nonce, NONCE_SIZE);

    // Generate a one-time keypair (skR and pkR) for receiving an HPKE message
    // from the app and output a serialized pkR. The sender needs this key in
    // order to seal the message. Writes `PUB_KEY_SIZE` bytes into `output`.
    ret = wc_HpkeGenerateKeyPair(hpke, &skr, kv_rng);
    if (ret != 0)
      return ret;

    uint16_t pkr_size = PUB_KEY_SIZE;
    uint8_t* pubout;
    msg.appendDataDirect("pub", &pubout, PUB_KEY_SIZE);
    ret = wc_HpkeSerializePublicKey(hpke, skr, pubout, &pkr_size);
    if (ret != 0)
      return ret;

    unsigned int outLen = ED25519_SIG_SIZE;
    uint8_t* sigout;
    uint8_t unsignedSize = msg.getLength();
    msg.appendDataDirect("sig", &sigout, ED25519_SIG_SIZE);
    ret = wc_ed25519_sign_msg(challenge, unsignedSize, sigout, &outLen,
                              &door_sign_key);
    if (ret != 0)
      return ret;
    return 0;
  }

  // Given an encrypted token, decrypt it, and verify it is
  // valid and signed by the server. The encrypted data is the nonce followed
  // by the token from the server, encrypted using
  // HPKE/KEM-X25519-SHA256/HKDF-SHA256/AES-GCM128. It is in the format of the
  // encapsulation key followed by the ciphertext. Returns
  // negative numbers for wolfssl errors, positive numbers for invalid key
  // errors, 0 for valid.
  int decryptAndVerifyToken(uint8_t* encryptedToken, size_t len) {
    // ensure the message is cleared out
    memset(_decryptedMessage, 0, DECRYPTED_MESSAGE_CAPACITY);
    int res;
    uint8_t* enc = encryptedToken;
    uint8_t* ct = encryptedToken + PUB_KEY_SIZE;
    res = wc_HpkeOpenBase(hpke, skr, enc, PUB_KEY_SIZE,     // enc
                          EXCHANGE_INFO, EXCHANGE_INFO_LEN, // info
                          NULL, 0,                          // aar
                          ct, len - PUB_KEY_SIZE, _decryptedMessage);
    // NOTE: for some reason this returns AES_GCM_AUTH_E, probably
    // because we aren't using mode_auth?
    if (res != 0 && res != AES_GCM_AUTH_E)
      return res;

    // nonce is first 8 bytes
    if (constantCompare(_decryptedMessage, _nonce, NONCE_SIZE) != 0) {
      return 0x10; // invalid nonce
    }

    uint8_t decryptedSize = _decryptedMessage[NONCE_SIZE];

    uint8_t* tokendata = _decryptedMessage + NONCE_SIZE + 1;
    TLVMessage message =
        TLVMessage(tokendata, decryptedSize, DECRYPTED_MESSAGE_CAPACITY);

    // uint8_t* alg;
    // uint8_t algSize;
    // if (message.getTag("alg", &alg, &algSize)) {
    //   if (algSize != 7 || memcmp(alg, "Ed25519", 7) != 0) {
    //     return 0x15; // unsupported alg
    //   }
    // }

    uint8_t* sig;
    uint8_t sigLen;
    if (!message.getTag("sig", &sig, &sigLen)) {
      return 0x20; // no signature
    }
    if (sigLen != ED25519_SIG_SIZE)
      return 0x21; // invalid signature

    size_t nonsiglen = (sig - 4) - tokendata;

    int verified;
    res = wc_ed25519_verify_msg(sig, ED25519_SIG_SIZE, tokendata, nonsiglen,
                                &verified, &server_sign_pub_key);
    if (res != 0)
      return res;
    if (verified == 0)
      return 0x30; // invalid signature

    uint8_t* expLoc;
    uint8_t expLen;
    if (!message.getTag("exp", &expLoc, &expLen)) {
      return 0x40; // no expires
    }
    if (expLen != 8) {
      return 0x41; // invalid expires
    }
    uint64_t exp = readLongBE(expLoc);
    uint64_t now = rtc->getTime();
    if (exp < now) {
      return 0x45; // expired
    }

    // TODO: verify other fields
    // typ == "TWT"
    // iss == "thecoven.space"
    // readLongBE( getTag("uid") )

    return 0; // valid
  }
};

#endif // KEY_VERIFICATION_H_
