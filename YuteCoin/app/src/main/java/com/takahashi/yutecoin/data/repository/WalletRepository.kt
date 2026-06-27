package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.WalletResponse

class WalletRepository {

    private val api = RetrofitClient.walletApi

    suspend fun getTransactions(address: String, page: Int = 1, limit: Int = 20): NetworkResult<WalletResponse> {
        return try {
            val response = api.getTransactions(address, page, limit)
            if (response.isSuccessful && response.body() != null) {
                NetworkResult.Success(response.body()!!)
            } else {
                NetworkResult.Error(response.code(), response.errorBody()?.string() ?: "Failed to load transactions")
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }
}
