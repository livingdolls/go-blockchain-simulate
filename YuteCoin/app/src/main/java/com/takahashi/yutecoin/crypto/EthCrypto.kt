package com.takahashi.yutecoin.crypto

import org.bouncycastle.jcajce.provider.digest.Keccak
import org.bouncycastle.math.ec.ECPoint
import java.math.BigInteger

object EthCrypto {
    private val keccak256 = Keccak.Digest256()

    fun keccak256Hash(data: ByteArray): ByteArray {
        return keccak256.digest(data)
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
        val hash = keccak256Hash(pubKeyBytes)
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

        val hash = keccak256Hash(pubKeyBytes)
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
        val addr = "0x" + toHex(addressBytes)
        val hash = keccak256Hash(addr.substring(2).toByteArray(Charsets.US_ASCII))
        val sb = StringBuilder("0x")
        for (i in addressBytes.indices) {
            val c = addressBytes[i].toInt() and 0xFF
            val h = hash[i].toInt() and 0xFF
            if (h >= 8) {
                sb.append(String.format("%02X", c))
            } else {
                sb.append(String.format("%02x", c))
            }
        }
        return sb.toString()
    }

    fun eip191Hash(message: String): ByteArray {
        val prefix = "\u0019Ethereum Signed Message:\n${message.length}$message"
        return keccak256Hash(prefix.toByteArray(Charsets.UTF_8))
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
