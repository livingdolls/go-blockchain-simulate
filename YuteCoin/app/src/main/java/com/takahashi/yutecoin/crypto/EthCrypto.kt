package com.takahashi.yutecoin.crypto

import org.bouncycastle.crypto.digests.KeccakDigest
import org.bouncycastle.math.ec.ECPoint
import java.math.BigInteger

object EthCrypto {

    fun keccak256Hash(data: ByteArray): ByteArray {
        val digest = KeccakDigest(256)
        val result = ByteArray(32)
        digest.update(data, 0, data.size)
        digest.doFinal(result, 0)
        return result
    }

    fun addressFromPoint(point: ECPoint): String {
        val x = point.affineXCoord.toBigInteger()
        val y = point.affineYCoord.toBigInteger()
        val pubKeyBytes = ByteArray(65)
        pubKeyBytes[0] = 0x04.toByte()
        val xPadded = Bip32.ser256(x)
        val yPadded = Bip32.ser256(y)
        System.arraycopy(xPadded, 0, pubKeyBytes, 1, 32)
        System.arraycopy(yPadded, 0, pubKeyBytes, 33, 32)
        // Ethereum address is derived from Keccak256 of the 64-byte public key
        // (x || y), NOT the 65-byte uncompressed format (0x04 || x || y).
        val hash = keccak256Hash(pubKeyBytes.copyOfRange(1, 65))
        val addressBytes = hash.copyOfRange(12, 32)
        return toChecksumAddress(addressBytes)
    }

    fun addressFromPublicKey(privateKey: BigInteger): String {
        val x = Secp256k1.publicKeyX(privateKey)
        val y = Secp256k1.publicKeyY(privateKey)

        // Uncompressed public key: 0x04 || x (32 bytes) || y (32 bytes)
        val pubKeyBytes = ByteArray(65)
        pubKeyBytes[0] = 0x04.toByte()
        val xPadded = Bip32.ser256(x)
        val yPadded = Bip32.ser256(y)
        System.arraycopy(xPadded, 0, pubKeyBytes, 1, 32)
        System.arraycopy(yPadded, 0, pubKeyBytes, 33, 32)

        // Ethereum address uses Keccak256 of the 64-byte public key (x || y).
        val hash = keccak256Hash(pubKeyBytes.copyOfRange(1, 65))
        val addressBytes = hash.copyOfRange(12, 32)
        return toChecksumAddress(addressBytes)
    }

    fun publicKeyToUncompressedHex(privateKey: BigInteger): String {
        val x = Secp256k1.publicKeyX(privateKey)
        val y = Secp256k1.publicKeyY(privateKey)
        val pubKeyBytes = ByteArray(65)
        pubKeyBytes[0] = 0x04.toByte()
        val xPadded = Bip32.ser256(x)
        val yPadded = Bip32.ser256(y)
        System.arraycopy(xPadded, 0, pubKeyBytes, 1, 32)
        System.arraycopy(yPadded, 0, pubKeyBytes, 33, 32)
        return "0x" + toHex(pubKeyBytes)
    }

    private fun recoverY(x: BigInteger): BigInteger {
        val curve = Secp256k1.DOMAIN.curve
        val p = curve.field.characteristic
        val a = curve.a.toBigInteger()
        val b = curve.b.toBigInteger()
        val xCubedPlusAxPlusB = x.pow(3).mod(p).add(a.multiply(x)).mod(p).add(b).mod(p)
        return xCubedPlusAxPlusB.modPow(p.add(BigInteger.ONE).divide(BigInteger.valueOf(4)), p)
    }

    fun toChecksumAddress(addressBytes: ByteArray): String {
        val addr = toHex(addressBytes)
        val hash = keccak256Hash(addr.toByteArray(Charsets.US_ASCII))
        val sb = StringBuilder("0x")
        for (i in addr.indices) {
            val ch = addr[i]
            val hashByte = hash[i / 2].toInt() and 0xFF
            val nibble = if (i % 2 == 0) hashByte shr 4 else hashByte and 0x0F
            if (nibble >= 8) {
                sb.append(ch.uppercaseChar())
            } else {
                sb.append(ch.lowercaseChar())
            }
        }
        return sb.toString()
    }

    fun eip191Hash(message: String): ByteArray {
        val messageBytes = message.toByteArray(Charsets.UTF_8)
        val prefix = "\u0019Ethereum Signed Message:\n${messageBytes.size}"
        val prefixBytes = prefix.toByteArray(Charsets.US_ASCII)
        val full = prefixBytes + messageBytes
        return keccak256Hash(full)
    }

    fun toHex(bytes: ByteArray): String {
        return bytes.joinToString("") { "%02x".format(it) }
    }

    fun toHexNoPrefix(bytes: ByteArray): String = toHex(bytes)

    fun hexToBytes(hex: String): ByteArray {
        val h = hex.removePrefix("0x")
        return (0 until h.length step 2).map {
            h.substring(it, it + 2).toInt(16).toByte()
        }.toByteArray()
    }

    fun cleanHexPrefix(hex: String): String = hex.removePrefix("0x")
}
