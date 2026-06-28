package com.takahashi.yutecoin.data.repository

import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.NotificationListResponse

class NotificationRepository {

    private val api = RetrofitClient.notificationApi

    suspend fun getNotifications(address: String, limit: Int = 20, offset: Int = 0): NetworkResult<NotificationListResponse> {
        return try {
            val response = api.getNotifications(address, limit, offset)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(response.body()!!.data!!)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun markAsRead(id: String): NetworkResult<Boolean> {
        return try {
            val response = api.markAsRead(id)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(true)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun markAllAsRead(address: String): NetworkResult<Boolean> {
        return try {
            val response = api.markAllAsRead(address)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(true)
            } else {
                NetworkResult.Error(response.code(), extractError(response.body()))
            }
        } catch (e: Exception) {
            NetworkResult.Error(0, e.message ?: "Network error")
        }
    }

    suspend fun delete(id: String): NetworkResult<Boolean> {
        return try {
            val response = api.delete(id)
            if (response.isSuccessful && response.body()?.success == true) {
                NetworkResult.Success(true)
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
