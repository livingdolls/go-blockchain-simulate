package com.takahashi.yutecoin.crypto

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Test
import java.math.BigInteger

class CryptoUnitTest {

    @Test
    fun keccak256_empty_matches_ethereum() {
        val emptyHash = EthCrypto.keccak256Hash(ByteArray(0))
        assertEquals(
            "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
            EthCrypto.toHex(emptyHash)
        )
    }

    @Test
    fun keccak256_abc_matches_ethereum() {
        val hash = EthCrypto.keccak256Hash("abc".toByteArray(Charsets.UTF_8))
        assertEquals(
            "4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45",
            EthCrypto.toHex(hash)
        )
    }

    @Test
    fun address_from_known_private_key() {
        val pk = BigInteger.ONE
        val x = Secp256k1.publicKeyX(pk)
        val y = Secp256k1.publicKeyY(pk)
        println("PUBLIC KEY for pk=1:")
        println("  x = ${x.toString(16)}")
        println("  y = ${y.toString(16)}")
        println("  expected x = 79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798")
        println("  expected y = 483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8")

        val pubKeyHex = EthCrypto.publicKeyToUncompressedHex(pk)
        println("  pubKeyHex = $pubKeyHex")
        val pubKeyBytes = EthCrypto.hexToBytes(pubKeyHex.removePrefix("0x"))
        val hash = EthCrypto.keccak256Hash(pubKeyBytes)
        println("  keccak(pubkey) = ${EthCrypto.toHex(hash)}")
        println("  last20 = ${EthCrypto.toHex(hash.copyOfRange(12, 32))}")

        val address = EthCrypto.addressFromPublicKey(pk)
        println("  address = $address")
        assertEquals("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf", address)
    }

    @Test
    fun eip191_hash_byte_length_for_ascii() {
        val msg = "Login to YuteBlockchain nonce:test-123"
        val expectedPrefix = "\u0019Ethereum Signed Message:\n${msg.toByteArray(Charsets.UTF_8).size}"
        val expected = EthCrypto.keccak256Hash(
            (expectedPrefix + msg).toByteArray(Charsets.UTF_8)
        )
        val actual = EthCrypto.eip191Hash(msg)
        assertEquals(EthCrypto.toHex(expected), EthCrypto.toHex(actual))
    }

    @Test
    fun sign_and_recover_challenge_message() {
        val privateKeyHex = "0x0000000000000000000000000000000000000000000000000000000000000001"
        val address = WalletGenerator.addressFromPrivateKey(privateKeyHex)
        val nonce = "test-nonce-123"
        val message = WalletSigner.buildChallengeMessage(nonce)

        val signature = WalletSigner.signChallengeMessage(nonce, privateKeyHex)
        val recovered = WalletSigner.recoverAddress(message, signature)

        assertNotNull(recovered)
        assertEquals(address.lowercase(), recovered!!.lowercase())
    }
}
