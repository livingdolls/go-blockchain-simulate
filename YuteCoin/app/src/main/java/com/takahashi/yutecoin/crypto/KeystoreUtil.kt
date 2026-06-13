package com.takahashi.yutecoin.crypto

import com.google.gson.Gson
import com.google.gson.JsonParser
import org.bouncycastle.crypto.generators.SCrypt
import org.bouncycastle.jcajce.provider.digest.Keccak
import java.security.SecureRandom
import javax.crypto.Cipher
import javax.crypto.spec.IvParameterSpec
import javax.crypto.spec.SecretKeySpec

object KeystoreUtil {

    // Scrypt parameters optimized for mobile (N=4096 vs desktop N=262144).
    // Lower N reduces memory usage from ~128MB to ~2MB, preventing OOM on Android.
    // The keystore JSON stores these params, so decryption works regardless.
    private const val SCRYPT_N = 4096
    private const val SCRYPT_R = 8
    private const val SCRYPT_P = 1
    private const val SCRYPT_DK_LEN = 32
    private const val CIPHER = "AES/CTR/NoPadding"
    private const val KEY_SIZE = 16

    private val gson = Gson()

    fun createKeystoreJson(privateKeyHex: String, password: String): String {
        val privateKeyBytes = EthCrypto.hexToBytes(EthCrypto.cleanHexPrefix(privateKeyHex))
        val salt = ByteArray(32).also { SecureRandom().nextBytes(it) }
        val iv = ByteArray(16).also { SecureRandom().nextBytes(it) }

        val derivedKey = SCrypt.generate(
            password.toByteArray(Charsets.UTF_8),
            salt,
            SCRYPT_N,
            SCRYPT_R,
            SCRYPT_P,
            SCRYPT_DK_LEN
        )

        val cipherKey = SecretKeySpec(derivedKey.copyOfRange(0, KEY_SIZE), "AES")
        val cipher = Cipher.getInstance(CIPHER)
        cipher.init(Cipher.ENCRYPT_MODE, cipherKey, IvParameterSpec(iv))
        val cipherText = cipher.doFinal(privateKeyBytes)

        val macKey = derivedKey.copyOfRange(KEY_SIZE, SCRYPT_DK_LEN)
        val macInput = macKey + cipherText
        val mac = keccak256(macInput)

        val keystore: Map<String, Any> = mapOf(
            "version" to 3,
            "id" to java.util.UUID.randomUUID().toString(),
            "crypto" to mapOf<String, Any>(
                "cipher" to "aes-128-ctr",
                "cipherparams" to mapOf("iv" to EthCrypto.toHexNoPrefix(iv)),
                "ciphertext" to EthCrypto.toHexNoPrefix(cipherText),
                "kdf" to "scrypt",
                "kdfparams" to mapOf(
                    "n" to SCRYPT_N,
                    "r" to SCRYPT_R,
                    "p" to SCRYPT_P,
                    "dklen" to SCRYPT_DK_LEN,
                    "salt" to EthCrypto.toHexNoPrefix(salt)
                ),
                "mac" to EthCrypto.toHexNoPrefix(mac)
            )
        )

        return gson.toJson(keystore)
    }

    fun decryptKeystore(json: String, password: String): WalletGenerator.WalletData {
        val root = JsonParser.parseString(json).asJsonObject
        val crypto = root["crypto"].asJsonObject

        val cipherText = EthCrypto.hexToBytes(crypto["ciphertext"].asString)
        val iv = EthCrypto.hexToBytes(crypto.getAsJsonObject("cipherparams")["iv"].asString)
        val expectedMac = EthCrypto.hexToBytes(crypto["mac"].asString)

        val kdfParams = crypto.getAsJsonObject("kdfparams")
        val salt = EthCrypto.hexToBytes(kdfParams["salt"].asString)
        val n = kdfParams["n"].asInt
        val r = kdfParams["r"].asInt
        val p = kdfParams["p"].asInt
        val dkLen = kdfParams["dklen"].asInt

        val derivedKey = SCrypt.generate(
            password.toByteArray(Charsets.UTF_8),
            salt,
            n,
            r,
            p,
            dkLen
        )

        val macKey = derivedKey.copyOfRange(KEY_SIZE, dkLen)
        val macInput = macKey + cipherText
        val computedMac = keccak256(macInput)
        if (!computedMac.contentEquals(expectedMac)) {
            throw SecurityException("Invalid password: MAC mismatch")
        }

        val cipherKey = SecretKeySpec(derivedKey.copyOfRange(0, KEY_SIZE), "AES")
        val cipher = Cipher.getInstance(CIPHER)
        cipher.init(Cipher.DECRYPT_MODE, cipherKey, IvParameterSpec(iv))
        val privateKeyBytes = cipher.doFinal(cipherText)

        val privateKeyHex = EthCrypto.toHexNoPrefix(privateKeyBytes)
        val address = WalletGenerator.addressFromPrivateKey(privateKeyHex)

        return WalletGenerator.WalletData(
            mnemonic = emptyList(),
            privateKeyHex = "0x$privateKeyHex",
            publicKeyHex = "",
            address = address
        )
    }

    private fun keccak256(input: ByteArray): ByteArray {
        return Keccak.Digest256().digest(input)
    }
}
