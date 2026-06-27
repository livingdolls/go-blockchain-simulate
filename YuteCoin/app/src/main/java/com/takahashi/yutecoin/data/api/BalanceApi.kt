package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.BalanceResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Path

interface BalanceApi {

    @GET("balance/{address}")
    suspend fun getBalance(@Path("address") address: String): Response<ApiResponse<BalanceResponse>>
}
