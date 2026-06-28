package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.BlockItem
import com.takahashi.yutecoin.data.dto.BlockStatsResponse
import com.takahashi.yutecoin.data.dto.NetworkResult

class BlockRepository {

    private val api = RetrofitClient.blockApi

    suspend fun getBlocks(limit: Int = 20, offset: Int = 0): NetworkResult<List<BlockItem>> {
        return try {
            val response = api.getBlocks(limit, offset)
            if (response.isSuccessful && response.body() != null) {
                NetworkResult.Success(response.body()!!.blocks ?: emptyList())
            } else {
                NetworkResult.Error(response.code(), "Failed to load blocks")
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun getStats(): NetworkResult<BlockStatsResponse> {
        return try {
            val response = api.getStats()
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
