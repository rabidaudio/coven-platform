package space.thecoven.android

import android.util.Log
import org.bouncycastle.crypto.hpke.HPKE
import java.security.KeyFactory
import java.security.Signature
import java.security.spec.X509EncodedKeySpec


class Authenticator(val doorSigningPubKey: ByteArray, val infoField: ByteArray) {

    // NOTE: when minSdk is 37 we can use the native Android libs for this instead of bouncycastle
    private val hpke = HPKE(
        HPKE.mode_base,
        HPKE.kem_X25519_SHA256,
        HPKE.kdf_HKDF_SHA256,
        HPKE.aead_AES_GCM128
    )

    @Throws(SecurityException::class)
    fun verifyChallengeAndEncryptToken(challenge: ByteArray, token: ByteArray): ByteArray {
        val challengeData = try {
            TLV.decode(challenge)
        } catch (e: IllegalArgumentException) {
            throw SecurityException("Invalid Challenge", e)
        }
        val signature = challengeData["sig"] ?: throw SecurityException("Invalid Challenge: sig")
        val signedMessage = challenge.copyOfRange(0, challenge.size - signature.size - 4)

        val keyspec = X509EncodedKeySpec(doorSigningPubKey)
        val key = KeyFactory.getInstance("Ed25519").generatePublic(keyspec)
        val sig = Signature.getInstance("Ed25519")
        sig.initVerify(key)
        sig.update(signedMessage)
        val valid = sig.verify(signature)
        if (!valid) {
            throw SecurityException("Invalid signature")
        }
        val nonce = challengeData["nce"] ?: throw SecurityException("Invalid Challenge: nce")
        val pubkey = challengeData["pub"] ?: throw SecurityException("Invalid Challenge: pub")

        Log.d("CRYPTO", "nonce=${nonce.toHexString()}")
        Log.d("CRYPTO", "pkR=${pubkey.toHexString()}")

        val plaintext = nonce + token

        val pkR = hpke.deserializePublicKey(pubkey)
        val aar = byteArrayOf() // empty
        val result = hpke.seal(pkR,
            infoField, aar,
            plaintext,
            null, null, // PSK (unused)
            null)

        val cipherText = result[0]
        val encapsulatedKey = result[1]
        Log.d("CRYPTO", "cipher=${cipherText.toHexString()}")
        Log.d("CRYPTO", "encapsulatedKey=${encapsulatedKey.toHexString()}")
        return encapsulatedKey+cipherText
    }
}
