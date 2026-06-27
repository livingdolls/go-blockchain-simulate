package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.BalanceResponse
import com.takahashi.yutecoin.data.dto.NetworkResult

class BalanceRepository {

    private val api = RetrofitClient.balanceApi

    suspend fun getBalance(address: String): NetworkResult<BalanceResponse> {
        return try {
            val response = api.getBalance(address)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!)
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
