package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.crypto.WalletGenerator
import com.takahashi.yutecoin.crypto.WalletSigner
import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.RegisterRequest

class AuthRepository {

    private val api = RetrofitClient.authApi

    suspend fun register(username: String, address: String, publicKey: String): NetworkResult<String> {
        return try {
            val response = api.register(RegisterRequest(username, address, publicKey))
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success("Registration successful")
            } else {
                val msg = response.body()?.message ?: "Registration failed"
                NetworkResult.Error(response.code(), msg)
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun loginWithMnemonic(mnemonic: List<String>, username: String): NetworkResult<String> {
        return try {
            val wallet = WalletGenerator.walletFromMnemonic(mnemonic)
            loginWithWallet(wallet.address, wallet.privateKeyHex, username)
        } catch (e: Exception) {
            NetworkResult.Error(0, "Invalid mnemonic: ${e.message}")
        }
    }

    suspend fun loginWithPrivateKey(privateKeyHex: String, username: String): NetworkResult<String> {
        return try {
            val address = WalletGenerator.addressFromPrivateKey(privateKeyHex)
            loginWithWallet(address, privateKeyHex, username)
        } catch (e: Exception) {
            NetworkResult.Error(0, "Invalid private key: ${e.message}")
        }
    }

    private suspend fun loginWithWallet(
        address: String,
        privateKeyHex: String,
        username: String
    ): NetworkResult<String> {
        // Step 1: Get challenge nonce
        val challengeResponse = try {
            api.getChallenge(address)
        } catch (e: Exception) {
            return NetworkResult.Error(0, "Failed to get challenge: ${e.message}")
        }

        if (!challengeResponse.isSuccessful || challengeResponse.body()?.success != true) {
            return NetworkResult.Error(
                challengeResponse.code(),
                challengeResponse.body()?.message ?: "Challenge request failed"
            )
        }

        val nonce = challengeResponse.body()!!.data!!.challenge

        // Step 2: Sign the challenge
        val signature = WalletSigner.signChallengeMessage(nonce, privateKeyHex)

        // Step 3: Verify
        val verifyResponse = try {
            api.verifyChallenge(
                com.takahashi.yutecoin.data.dto.VerifyRequest(
                    address = address,
                    signature = signature,
                    nonce = nonce,
                    username = username
                )
            )
        } catch (e: Exception) {
            return NetworkResult.Error(0, "Verification failed: ${e.message}")
        }

        return if (verifyResponse.isSuccessful && verifyResponse.body()?.success == true) {
            NetworkResult.Success(address)
        } else {
            NetworkResult.Error(
                verifyResponse.code(),
                verifyResponse.body()?.message ?: "Login failed"
            )
        }
    }

    suspend fun fetchProfile(): NetworkResult<String> {
        return try {
            val response = api.getProfile()
            if (response.isSuccessful && response.body()?.success == true) {
                val user = response.body()!!.data!!
                NetworkResult.Success(user.address)
            } else {
                NetworkResult.Error(response.code(), response.body()?.message ?: "Not authenticated")
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    fun isLoggedIn(): Boolean {
        return RetrofitClient.cookieJar.hasAuthCookie()
    }

    fun logout() {
        RetrofitClient.clearSession()
    }
}
