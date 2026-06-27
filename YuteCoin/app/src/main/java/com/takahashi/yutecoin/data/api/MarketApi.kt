package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.CandleResponse
import com.takahashi.yutecoin.data.dto.MarketResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Query

interface MarketApi {

    @GET("market")
    suspend fun getMarketState(): Response<ApiResponse<MarketResponse>>

    @GET("candles")
    suspend fun getCandles(
        @Query("interval") interval: String = "1m",
        @Query("limit") limit: Int = 50
    ): Response<ApiResponse<List<CandleResponse>>>
}
