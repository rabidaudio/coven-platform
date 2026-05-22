package space.thecoven.android

import android.content.Context
import be.appmire.flutterkeychain.AesStringEncryptor
import be.appmire.flutterkeychain.RsaKeyStoreKeyWrapper
import be.appmire.flutterkeychain.StringEncryptor
import androidx.core.content.edit

/**
 * This class works by calling the classes underlying `flutter_keychain` directly to
 * access the saved user token. We could instead request it from the frontend but
 * this seems less error-prone.
 */
class TokenManager(context: Context) {
    companion object {
        const val USER_TOKEN_KEY = "user_token"
    }

    private val preferences = context.getSharedPreferences("FlutterKeychain", Context.MODE_PRIVATE)
    private val encryptor = AesStringEncryptor(
        preferences = preferences,
        keyWrapper = RsaKeyStoreKeyWrapper(context)
    ) as StringEncryptor

    fun getKey(key: String): String? {
        return encryptor.decrypt(preferences.getString(key, null))
    }

    fun getToken(): String? = getKey(USER_TOKEN_KEY)

    fun clearToken() {
        preferences.edit(commit = true) { remove(USER_TOKEN_KEY) }
    }
}
