package com.takahashi.yutecoin.crypto

import android.content.Context
import android.util.Log
import org.bouncycastle.crypto.digests.SHA256Digest
import org.bouncycastle.crypto.digests.SHA512Digest
import org.bouncycastle.crypto.generators.PKCS5S2ParametersGenerator
import org.bouncycastle.crypto.params.KeyParameter
import java.security.SecureRandom

object Bip39 {

    private const val TAG = "Bip39"

    // Shared SecureRandom — lazy initialized
    private val secureRandom: SecureRandom by lazy {
        Log.d(TAG, "Initializing SecureRandom...")
        val start = System.currentTimeMillis()
        val sr = SecureRandom()
        Log.d(TAG, "SecureRandom ready in ${System.currentTimeMillis() - start}ms")
        sr
    }

    // Wordlist — loaded from assets, lazy init
    @Volatile
    private var _words: List<String>? = null

    fun loadFromAssets(context: Context) {
        if (_words != null) return
        synchronized(this) {
            if (_words != null) return
            val start = System.currentTimeMillis()
            val words = context.assets.open("bip39_english.txt").bufferedReader().useLines { lines ->
                lines.map { it.trim() }.filter { it.isNotEmpty() }.toList()
            }
            Log.d(TAG, "Wordlist loaded from assets: ${words.size} words in ${System.currentTimeMillis() - start}ms")
            require(words.size == 2048) { "BIP39 wordlist must have 2048 words, got ${words.size}" }
            _words = words
        }
    }

    val englishWords: List<String>
        get() = _words ?: throw IllegalStateException(
            "Bip39 not initialized. Call Bip39.loadFromAssets(context) in Application.onCreate() first."
        )

    fun generateMnemonic(): List<String> {
        Log.d(TAG, "generateMnemonic: start")
        val entropy = ByteArray(16)
        secureRandom.nextBytes(entropy)
        Log.d(TAG, "generateMnemonic: entropy ready")
        val result = mnemonicFromEntropy(entropy)
        Log.d(TAG, "generateMnemonic: done, ${result.size} words")
        return result
    }

    fun mnemonicFromEntropy(entropy: ByteArray): List<String> {
        require(entropy.size in listOf(16, 20, 24, 28, 32)) { "Entropy must be 16/20/24/28/32 bytes" }
        val words = englishWords

        val hash = sha256(entropy)
        val checksumBits = entropy.size * 8 / 32
        val totalBits = entropy.size * 8 + checksumBits

        val bits = BooleanArray(totalBits)
        for (i in entropy.indices) {
            val b = entropy[i].toInt() and 0xFF
            for (j in 0..7) {
                bits[i * 8 + j] = (b ushr (7 - j)) and 1 == 1
            }
        }
        for (i in 0 until checksumBits) {
            bits[entropy.size * 8 + i] = ((hash[0].toInt() and 0xFF) ushr (7 - i)) and 1 == 1
        }

        val result = mutableListOf<String>()
        for (i in 0 until totalBits / 11) {
            var index = 0
            for (j in 0..10) {
                if (bits[i * 11 + j]) index = index or (1 shl (10 - j))
            }
            require(index in 0..2047) { "Word index $index out of bounds" }
            result.add(words[index])
        }
        return result
    }

    fun validateMnemonic(mnemonic: List<String>): Boolean {
        if (mnemonic.size !in listOf(12, 15, 18, 21, 24)) return false
        return try {
            mnemonic.all { it in englishWords } && mnemonicFromEntropy(reconstructEntropy(mnemonic)) == mnemonic
        } catch (_: Exception) {
            false
        }
    }

    private fun reconstructEntropy(mnemonic: List<String>): ByteArray {
        val words = englishWords
        val totalBits = mnemonic.size * 11
        val checksumBits = totalBits / 33
        val entropyBits = totalBits - checksumBits
        val entropyBytes = (entropyBits + 7) / 8
        val bits = BooleanArray(totalBits)
        for (i in mnemonic.indices) {
            val idx = words.indexOf(mnemonic[i])
            for (j in 0..10) bits[i * 11 + j] = (idx ushr (10 - j)) and 1 == 1
        }
        val entropy = ByteArray(entropyBytes)
        for (i in 0 until entropyBits) {
            if (bits[i]) entropy[i / 8] = (entropy[i / 8].toInt() or (1 shl (7 - (i % 8)))).toByte()
        }
        return entropy
    }

    fun seedFromMnemonic(mnemonic: List<String>, passphrase: String = ""): ByteArray {
        Log.d(TAG, "seedFromMnemonic: start, ${mnemonic.size} words")
        val mnemonicBytes = mnemonic.joinToString(" ").toByteArray(Charsets.UTF_8)
        val salt = ("mnemonic" + passphrase).toByteArray(Charsets.UTF_8)
        val generator = PKCS5S2ParametersGenerator(SHA512Digest())
        generator.init(mnemonicBytes, salt, 2048)
        val params = generator.generateDerivedParameters(512)
        val key = (params as KeyParameter).key
        Log.d(TAG, "seedFromMnemonic: done, seed=${key.size} bytes")
        return key
    }

    private fun sha256(data: ByteArray): ByteArray {
        val digest = SHA256Digest()
        val result = ByteArray(digest.digestSize)
        digest.update(data, 0, data.size)
        digest.doFinal(result, 0)
        return result
    }
}
