package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.BuySellRequest
import com.takahashi.yutecoin.data.dto.NonceResponse
import com.takahashi.yutecoin.data.dto.SendRequest
import com.takahashi.yutecoin.data.dto.SendResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path

interface TransactionApi {

    @GET("generate-tx-nonce/{address}")
    suspend fun generateNonce(@Path("address") address: String): Response<NonceResponse>

    @POST("transaction/send")
    suspend fun send(@Body request: SendRequest): Response<ApiResponse<SendResponse>>

    @POST("transaction/buy")
    suspend fun buy(@Body request: BuySellRequest): Response<ApiResponse<SendResponse>>

    @POST("transaction/sell")
    suspend fun sell(@Body request: BuySellRequest): Response<ApiResponse<SendResponse>>
}
