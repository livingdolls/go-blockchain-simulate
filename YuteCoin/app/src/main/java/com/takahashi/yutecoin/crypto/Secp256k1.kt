package com.takahashi.yutecoin.crypto

import org.bouncycastle.asn1.sec.SECNamedCurves
import org.bouncycastle.asn1.x9.X9ECParameters
import org.bouncycastle.crypto.digests.SHA256Digest
import org.bouncycastle.crypto.params.ECDomainParameters
import org.bouncycastle.crypto.params.ECPrivateKeyParameters
import org.bouncycastle.crypto.signers.ECDSASigner
import org.bouncycastle.crypto.signers.HMacDSAKCalculator
import org.bouncycastle.math.ec.ECPoint
import java.math.BigInteger
import java.security.SecureRandom

object Secp256k1 {
    private val CURVE_PARAMS: X9ECParameters = SECNamedCurves.getByName("secp256k1")
    val DOMAIN: ECDomainParameters = ECDomainParameters(
        CURVE_PARAMS.curve, CURVE_PARAMS.g, CURVE_PARAMS.n, CURVE_PARAMS.h
    )
    val N: BigInteger = DOMAIN.n

    fun publicKeyFromPrivate(privateKey: BigInteger): ECPoint {
        return DOMAIN.g.multiply(privateKey).normalize()
    }

    /**
     * Returns the x-coordinate of the public key corresponding to the private key.
     */
    fun publicKeyX(privateKey: BigInteger): BigInteger {
        return publicKeyFromPrivate(privateKey).affineXCoord.toBigInteger()
    }

    /**
     * Returns the y-coordinate of the public key corresponding to the private key.
     */
    fun publicKeyY(privateKey: BigInteger): BigInteger {
        return publicKeyFromPrivate(privateKey).affineYCoord.toBigInteger()
    }

    fun sign(hash: ByteArray, privateKey: BigInteger): ECDSASignature {
        val signer = ECDSASigner(HMacDSAKCalculator(SHA256Digest()))
        signer.init(true, ECPrivateKeyParameters(privateKey, DOMAIN))
        val components = signer.generateSignature(hash)
        var r = components[0] as BigInteger
        var s = components[1] as BigInteger

        // Enforce low-s (EIP-2)
        val halfN = N.shiftRight(1)
        if (s > halfN) s = N.subtract(s)

        // Determine recovery id v
        var recId = -1
        val expectedX = publicKeyX(privateKey)
        val expectedY = publicKeyY(privateKey)
        for (i in 0..3) {
            try {
                val recovered = recoverFromSignature(i, r, s, hash)
                if (recovered.affineXCoord.toBigInteger() == expectedX &&
                    recovered.affineYCoord.toBigInteger() == expectedY) {
                    recId = i
                    break
                }
            } catch (_: Exception) {}
        }
        require(recId >= 0) { "Failed to determine recovery id" }
        return ECDSASignature(r, s, recId)
    }

    fun recoverFromSignature(recId: Int, r: BigInteger, s: BigInteger, hash: ByteArray): ECPoint {
        require(recId in 0..3) { "recId must be 0-3" }
        val n = N
        val i = BigInteger.valueOf(recId.toLong() / 2)
        val x = r.add(i.multiply(n))
        require(x < DOMAIN.curve.field.characteristic) { "x too large" }

        // Recover y from x on the curve
        val curve = DOMAIN.curve
        val a = curve.a.toBigInteger()
        val b = curve.b.toBigInteger()
        val p = curve.field.characteristic

        val y2 = x.pow(3).mod(p).add(a.multiply(x)).mod(p).add(b).mod(p)
        var y = y2.modPow(p.add(BigInteger.ONE).divide(BigInteger.valueOf(4)), p)

        val isYOdd = y.testBit(0)
        val shouldBeOdd = (recId and 1) == 1
        if (isYOdd != shouldBeOdd) {
            y = p.subtract(y)
        }

        // Reconstruct R point
        val R = curve.createPoint(x, y)
        val rInv = r.modInverse(n)
        val z = BigInteger(1, hash)

        // Q = r^-1 * (s*R - z*G)
        val u1 = s.multiply(rInv).mod(n)       // s * r^(-1)
        val u2 = n.subtract(z).multiply(rInv).mod(n)  // (n - z) * r^(-1)
        val Q = R.multiply(u1).add(DOMAIN.g.multiply(u2)).normalize()
        return Q
    }

    fun generatePrivateKey(): BigInteger {
        val rand = SecureRandom()
        var key: BigInteger
        do {
            key = BigInteger(N.bitLength(), rand)
        } while (key >= N || key == BigInteger.ZERO)
        return key
    }

    data class ECDSASignature(val r: BigInteger, val s: BigInteger, val v: Int)

    // Convert an ECPoint's x-coordinate to a BigInteger
    fun ECPoint.xBigInt() = this.affineXCoord.toBigInteger()
}
