package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.NotificationListResponse
import retrofit2.Response
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.PUT
import retrofit2.http.Path
import retrofit2.http.Query

interface NotificationApi {

    @GET("notifications")
    suspend fun getNotifications(
        @Query("address") address: String,
        @Query("limit") limit: Int = 20,
        @Query("offset") offset: Int = 0
    ): Response<ApiResponse<NotificationListResponse>>

    @PUT("notifications/{id}/read")
    suspend fun markAsRead(@Path("id") id: String): Response<ApiResponse<Any>>

    @PUT("notifications/read-all")
    suspend fun markAllAsRead(@Query("address") address: String): Response<ApiResponse<Any>>

    @DELETE("notifications/{id}")
    suspend fun delete(@Path("id") id: String): Response<ApiResponse<Any>>
}
