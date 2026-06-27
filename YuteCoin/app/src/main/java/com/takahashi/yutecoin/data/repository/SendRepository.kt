package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.BuySellRequest
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.SendRequest

class SendRepository {

    private val api = RetrofitClient.transactionApi

    suspend fun generateNonce(address: String): NetworkResult<String> {
        return try {
            val response = api.generateNonce(address)
            if (response.isSuccessful && response.body() != null) {
                NetworkResult.Success(response.body()!!.nonce)
            } else {
                NetworkResult.Error(response.code(), "Failed to generate nonce")
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun send(request: SendRequest): NetworkResult<String> {
        return try {
            val response = api.send(request)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!.message)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun buy(request: BuySellRequest): NetworkResult<String> {
        return try {
            val response = api.buy(request)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!.message)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun sell(request: BuySellRequest): NetworkResult<String> {
        return try {
            val response = api.sell(request)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!.message)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    private fun extractError(body: ApiResponse<*>?): String {
        if (body == null) return "Unknown error"
        return body.error ?: "Request failed (code ${body.code ?: "?"})"
    }
}
