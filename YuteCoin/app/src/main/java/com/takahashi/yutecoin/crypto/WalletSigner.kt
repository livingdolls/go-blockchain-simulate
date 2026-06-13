package com.takahashi.yutecoin.crypto

import java.math.BigInteger

object WalletSigner {

    fun signChallengeMessage(nonce: String, privateKeyHex: String): String {
        val message = buildChallengeMessage(nonce)
        return WalletGenerator.signMessageToHex(message, privateKeyHex)
    }

    fun buildChallengeMessage(nonce: String): String {
        return "Login to YuteBlockchain nonce:$nonce"
    }

    fun recoverAddress(message: String, signatureHex: String): String? {
        return try {
            val hash = EthCrypto.eip191Hash(message)
            val sig = parseSignature(signatureHex)
            val point = Secp256k1.recoverFromSignature(sig.v - 27, sig.r, sig.s, hash)
            EthCrypto.addressFromPoint(point)
        } catch (e: Exception) {
            null
        }
    }

    fun parseSignature(signatureHex: String): Secp256k1.ECDSASignature {
        val hex = signatureHex.removePrefix("0x")
        val bytes = EthCrypto.hexToBytes(hex)
        require(bytes.size == 65) { "Signature must be 65 bytes" }
        val r = BigInteger(1, bytes.copyOfRange(0, 32))
        val s = BigInteger(1, bytes.copyOfRange(32, 64))
        var v = bytes[64].toInt() and 0xFF
        if (v < 27) v += 27
        return Secp256k1.ECDSASignature(r, s, v)
    }
}
