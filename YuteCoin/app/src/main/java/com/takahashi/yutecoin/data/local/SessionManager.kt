package com.takahashi.yutecoin.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

class SessionManager(context: Context) {

    companion object {
        private const val PREFS_NAME = "yutecoin_secure"
        private const val KEY_MNEMONIC = "wallet_mnemonic"
        private const val KEY_PRIVATE_KEY = "wallet_private_key"
        private const val KEY_ADDRESS = "wallet_address"
        private const val KEY_USERNAME = "username"
        private const val KEY_IS_LOGGED_IN = "is_logged_in"
    }

    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val securePrefs: SharedPreferences = EncryptedSharedPreferences.create(
        context,
        PREFS_NAME,
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    fun saveWallet(mnemonic: List<String>, privateKeyHex: String, address: String) {
        securePrefs.edit()
            .putString(KEY_MNEMONIC, mnemonic.joinToString(" "))
            .putString(KEY_PRIVATE_KEY, privateKeyHex)
            .putString(KEY_ADDRESS, address)
            .apply()
    }

    fun saveMnemonic(mnemonic: List<String>, username: String) {
        securePrefs.edit()
            .putString(KEY_MNEMONIC, mnemonic.joinToString(" "))
            .putString(KEY_USERNAME, username)
            .apply()
    }

    fun saveUsername(username: String) {
        securePrefs.edit().putString(KEY_USERNAME, username).apply()
    }

    fun getUsername(): String? = securePrefs.getString(KEY_USERNAME, null)

    fun getMnemonic(): List<String>? {
        val mnemonicStr = securePrefs.getString(KEY_MNEMONIC, null) ?: return null
        return mnemonicStr.split(" ").takeIf { it.size >= 12 }
    }

    fun getPrivateKeyHex(): String? = securePrefs.getString(KEY_PRIVATE_KEY, null)

    fun getAddress(): String? = securePrefs.getString(KEY_ADDRESS, null)

    fun isLoggedIn(): Boolean = securePrefs.getBoolean(KEY_IS_LOGGED_IN, false)

    fun setLoggedIn(loggedIn: Boolean) {
        securePrefs.edit().putBoolean(KEY_IS_LOGGED_IN, loggedIn).apply()
    }

    fun clearAll() {
        securePrefs.edit().clear().apply()
    }
}
