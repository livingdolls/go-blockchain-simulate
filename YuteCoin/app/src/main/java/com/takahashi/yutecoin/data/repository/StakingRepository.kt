package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.StakeRequest
import com.takahashi.yutecoin.data.dto.StakeResponse
import com.takahashi.yutecoin.data.dto.StakingInfoResponse
import com.takahashi.yutecoin.data.dto.StakingStatusResponse
import com.takahashi.yutecoin.data.dto.UnstakeRequest
import com.takahashi.yutecoin.data.dto.UnstakeResponse

class StakingRepository {

    private val api = RetrofitClient.stakingApi

    suspend fun stake(request: StakeRequest): NetworkResult<StakeResponse> {
        return try {
            val response = api.stake(request)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun unstake(request: UnstakeRequest): NetworkResult<UnstakeResponse> {
        return try {
            val response = api.unstake(request)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun getStatus(address: String): NetworkResult<StakingStatusResponse> {
        return try {
            val response = api.getStatus(address)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun getInfo(): NetworkResult<StakingInfoResponse> {
        return try {
            val response = api.getInfo()
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
