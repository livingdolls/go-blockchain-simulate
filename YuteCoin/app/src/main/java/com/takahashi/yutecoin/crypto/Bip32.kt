package com.takahashi.yutecoin.crypto

import org.bouncycastle.crypto.digests.RIPEMD160Digest
import org.bouncycastle.crypto.digests.SHA256Digest
import org.bouncycastle.crypto.digests.SHA512Digest
import org.bouncycastle.crypto.macs.HMac
import org.bouncycastle.crypto.params.KeyParameter
import org.bouncycastle.math.ec.ECPoint
import java.math.BigInteger

object Bip32 {
    private const val HARDENED_OFFSET = 0x80000000.toInt()
    private val CURVE_N = Secp256k1.N

    data class ExtendedKey(
        val privateKey: BigInteger,
        val chainCode: ByteArray,
        val depth: Int = 0,
        val parentFingerprint: Int = 0,
        val childNumber: Int = 0
    )

    fun masterKey(seed: ByteArray): ExtendedKey {
        val hmac = hmacSha512("Bitcoin seed".toByteArray(Charsets.UTF_8), seed)
        val privateKey = BigInteger(1, hmac.copyOfRange(0, 32))
        require(privateKey > BigInteger.ZERO && privateKey < CURVE_N) {
            "Invalid master key"
        }
        val chainCode = hmac.copyOfRange(32, 64)
        return ExtendedKey(privateKey, chainCode)
    }

    fun deriveHardened(parent: ExtendedKey, index: Int): ExtendedKey {
        val i = index or HARDENED_OFFSET
        val data = ByteArray(37)
        data[0] = 0
        System.arraycopy(ser256(parent.privateKey), 0, data, 1, 32)
        ser32(i).copyInto(data, 33)
        return ckdPriv(parent, data, i)
    }

    fun deriveNonHardened(parent: ExtendedKey, index: Int): ExtendedKey {
        // Serialize compressed public key: 33 bytes (0x02/0x03 || x)
        val pubPoint = Secp256k1.publicKeyFromPrivate(parent.privateKey)
        val compressed = compressedPublicKey(pubPoint)
        val data = ByteArray(33 + 4)
        System.arraycopy(compressed, 0, data, 0, 33)
        ser32(index).copyInto(data, 33)
        return ckdPriv(parent, data, index)
    }

    private fun ckdPriv(parent: ExtendedKey, data: ByteArray, childNumber: Int): ExtendedKey {
        val i = hmacSha512(parent.chainCode, data)
        val il = BigInteger(1, i.copyOfRange(0, 32))
        val ir = i.copyOfRange(32, 64)
        require(il < CURVE_N) { "Invalid derived key" }
        val childPrivateKey = il.add(parent.privateKey).mod(CURVE_N)
        require(childPrivateKey != BigInteger.ZERO) { "Invalid child key" }

        val parentPubPoint = Secp256k1.publicKeyFromPrivate(parent.privateKey)
        val fp = fingerprint(parentPubPoint)
        return ExtendedKey(childPrivateKey, ir, parent.depth + 1, fp, childNumber)
    }

    fun compressedPublicKey(point: ECPoint): ByteArray {
        val x = point.affineXCoord.toBigInteger()
        val yIsOdd = point.affineYCoord.toBigInteger().testBit(0)
        val prefix = if (yIsOdd) 0x03.toByte() else 0x02.toByte()
        return byteArrayOf(prefix) + ser256(x)
    }

    fun fingerprint(point: ECPoint): Int {
        val compressed = compressedPublicKey(point)
        val sha256 = SHA256Digest()
        val shaHash = ByteArray(sha256.digestSize)
        sha256.update(compressed, 0, compressed.size)
        sha256.doFinal(shaHash, 0)

        val ripe = RIPEMD160Digest()
        val ripeHash = ByteArray(ripe.digestSize)
        ripe.update(shaHash, 0, shaHash.size)
        ripe.doFinal(ripeHash, 0)
        return (ripeHash[0].toInt() and 0xFF shl 24) or
                (ripeHash[1].toInt() and 0xFF shl 16) or
                (ripeHash[2].toInt() and 0xFF shl 8) or
                (ripeHash[3].toInt() and 0xFF)
    }

    fun ser256(x: BigInteger): ByteArray {
        val bytes = x.toByteArray()
        if (bytes.size > 32) return bytes.copyOfRange(bytes.size - 32, bytes.size)
        if (bytes.size < 32) {
            val padded = ByteArray(32)
            System.arraycopy(bytes, 0, padded, 32 - bytes.size, bytes.size)
            return padded
        }
        return bytes
    }

    fun ser32(i: Int): ByteArray {
        return byteArrayOf(
            (i shr 24).toByte(), (i shr 16).toByte(), (i shr 8).toByte(), i.toByte()
        )
    }

    private fun hmacSha512(key: ByteArray, data: ByteArray): ByteArray {
        val hmac = HMac(SHA512Digest())
        hmac.init(KeyParameter(key))
        val result = ByteArray(hmac.macSize)
        hmac.update(data, 0, data.size)
        hmac.doFinal(result, 0)
        return result
    }
}
