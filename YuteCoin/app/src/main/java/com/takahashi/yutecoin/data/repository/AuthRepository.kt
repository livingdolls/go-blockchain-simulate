package com.takahashi.yutecoin.data.repository

import android.util.Log
import com.takahashi.yutecoin.crypto.WalletGenerator
import com.takahashi.yutecoin.crypto.WalletSigner
import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.RegisterRequest

class AuthRepository {

    private val api = RetrofitClient.authApi

    suspend fun register(username: String, address: String, publicKey: String): NetworkResult<String> {
        return try {
            val response = api.register(RegisterRequest(username, address, publicKey))
            Log.d("AuthRepo", "register response: code=${response.code()}, success=${response.body()?.success}")
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success("Registration successful")
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            Log.e("AuthRepo", "register exception", e)
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun loginWithMnemonic(mnemonic: List<String>, username: String): NetworkResult<String> {
        return try {
            val wallet = WalletGenerator.walletFromMnemonic(mnemonic)
            loginWithWallet(wallet.address, wallet.privateKeyHex, username)
        } catch (e: Exception) {
            Log.e("AuthRepo", "loginWithMnemonic exception", e)
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
                extractError(challengeResponse.body())
            )
        }

        val nonce = challengeResponse.body()!!.data!!.challenge
        Log.d("AuthRepo", "Got challenge nonce: $nonce")

        // Step 2: Sign the challenge
        val signature = WalletSigner.signChallengeMessage(nonce, privateKeyHex)
        Log.d("AuthRepo", "Signed challenge, signature length: ${signature.length}")

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

        Log.d("AuthRepo", "verify response: code=${verifyResponse.code()}, success=${verifyResponse.body()?.success}")

        return if (verifyResponse.isSuccessful && verifyResponse.body()?.success == true) {
            NetworkResult.Success(address)
        } else {
            NetworkResult.Error(
                verifyResponse.code(),
                extractError(verifyResponse.body())
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
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    private fun extractError(body: ApiResponse<*>?): String {
        if (body == null) return "Unknown error"
        val parts = mutableListOf<String>()
        body.error?.let { parts.add(it) }
        body.field?.let { f ->
            // Find field-specific errors in details
            body.details
                ?.filter { it.field == f }
                ?.forEach { parts.add("${it.field}: ${it.message}") }
        }
        body.details
            ?.filter { body.field == null || it.field != body.field }
            ?.forEach { parts.add("${it.field}: ${it.message}") }
        return parts.joinToString("; ").ifEmpty { "Request failed (code ${body.code ?: "?"})" }
    }

    fun isLoggedIn(): Boolean {
        return RetrofitClient.cookieJar.hasAuthCookie()
    }

    fun logout() {
        RetrofitClient.clearSession()
    }
}
