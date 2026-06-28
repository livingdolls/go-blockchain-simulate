package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.BlockListResponse
import com.takahashi.yutecoin.data.dto.BlockStatsResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Query

interface BlockApi {

    @GET("blocks")
    suspend fun getBlocks(
        @Query("limit") limit: Int = 20,
        @Query("offset") offset: Int = 0
    ): Response<BlockListResponse>

    @GET("blocks/stats")
    suspend fun getStats(): Response<ApiResponse<BlockStatsResponse>>
}
