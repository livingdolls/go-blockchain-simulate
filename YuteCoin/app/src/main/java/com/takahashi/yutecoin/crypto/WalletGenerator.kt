package com.takahashi.yutecoin.crypto

import android.util.Log
import java.math.BigInteger

object WalletGenerator {

    data class WalletData(
        val mnemonic: List<String>,
        val privateKeyHex: String,
        val publicKeyHex: String,
        val address: String
    )

    fun generateMnemonic(): List<String> {
        Log.d("Wallet", "generateMnemonic: start")
        val result = Bip39.generateMnemonic()
        Log.d("Wallet", "generateMnemonic: done, words=${result.size}")
        return result
    }

    fun generateWallet(): WalletData {
        Log.d("Wallet", "generateWallet: start")
        val mnemonic = generateMnemonic()
        Log.d("Wallet", "generateWallet: mnemonic done, generating seed")
        val seed = Bip39.seedFromMnemonic(mnemonic)
        Log.d("Wallet", "generateWallet: seed done (${seed.size} bytes), deriving master")
        val masterKey = Bip32.masterKey(seed)
        Log.d("Wallet", "generateWallet: master key done, deriving BIP44")
        val derived = deriveBip44(masterKey)
        Log.d("Wallet", "generateWallet: BIP44 done, building data")
        val result = buildWalletData(mnemonic, derived)
        Log.d("Wallet", "generateWallet: done, address=${result.address}")
        return result
    }

    fun walletFromMnemonic(mnemonic: List<String>): WalletData {
        val seed = Bip39.seedFromMnemonic(mnemonic)
        val masterKey = Bip32.masterKey(seed)
        val derived = deriveBip44(masterKey)
        return buildWalletData(mnemonic, derived)
    }

    fun isValidMnemonic(mnemonic: List<String>): Boolean = Bip39.validateMnemonic(mnemonic)

    fun addressFromPrivateKey(privateKeyHex: String): String {
        val pk = BigInteger(1, EthCrypto.hexToBytes(privateKeyHex))
        return EthCrypto.addressFromPublicKey(pk)
    }

    fun signMessageToHex(message: String, privateKeyHex: String): String {
        val pk = BigInteger(1, EthCrypto.hexToBytes(privateKeyHex))
        val hash = EthCrypto.eip191Hash(message)
        val sig = Secp256k1.sign(hash, pk)
        val r = Bip32.ser256(sig.r)
        val s = Bip32.ser256(sig.s)
        val v = sig.v + 27 // EIP-191: v = recId + 27
        val rs = r + s + byteArrayOf(v.toByte())
        return "0x" + EthCrypto.toHex(rs)
    }

    private fun deriveBip44(masterKey: Bip32.ExtendedKey): Bip32.ExtendedKey {
        // m/44'/60'/0'/0/0
        val purpose = Bip32.deriveHardened(masterKey, 44)
        val coinType = Bip32.deriveHardened(purpose, 60)
        val account = Bip32.deriveHardened(coinType, 0)
        val change = Bip32.deriveNonHardened(account, 0)
        return Bip32.deriveNonHardened(change, 0)
    }

    private fun buildWalletData(mnemonic: List<String>, key: Bip32.ExtendedKey): WalletData {
        return WalletData(
            mnemonic = mnemonic,
            privateKeyHex = "0x" + EthCrypto.toHex(Bip32.ser256(key.privateKey)),
            publicKeyHex = EthCrypto.publicKeyToUncompressedHex(key.privateKey),
            address = EthCrypto.addressFromPublicKey(key.privateKey)
        )
    }
}
