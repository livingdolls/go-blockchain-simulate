package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.StakeRequest
import com.takahashi.yutecoin.data.dto.StakeResponse
import com.takahashi.yutecoin.data.dto.StakingInfoResponse
import com.takahashi.yutecoin.data.dto.StakingStatusResponse
import com.takahashi.yutecoin.data.dto.UnstakeRequest
import com.takahashi.yutecoin.data.dto.UnstakeResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Query

interface StakingApi {

    @POST("staking/stake")
    suspend fun stake(@Body request: StakeRequest): Response<ApiResponse<StakeResponse>>

    @POST("staking/unstake")
    suspend fun unstake(@Body request: UnstakeRequest): Response<ApiResponse<UnstakeResponse>>

    @GET("staking/status")
    suspend fun getStatus(@Query("address") address: String): Response<ApiResponse<StakingStatusResponse>>

    @GET("staking/info")
    suspend fun getInfo(): Response<ApiResponse<StakingInfoResponse>>
}
