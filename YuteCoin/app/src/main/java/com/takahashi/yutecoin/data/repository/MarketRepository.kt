package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.CandleResponse
import com.takahashi.yutecoin.data.dto.MarketResponse
import com.takahashi.yutecoin.data.dto.NetworkResult

class MarketRepository {

    private val api = RetrofitClient.marketApi

    suspend fun getMarketState(): NetworkResult<MarketResponse> {
        return try {
            val response = api.getMarketState()
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun getCandles(interval: String = "1m", limit: Int = 50): NetworkResult<List<CandleResponse>> {
        return try {
            val response = api.getCandles(interval, limit)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data ?: emptyList())
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
