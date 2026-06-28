package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.PortfolioResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Path

interface PortfolioApi {

    @GET("portfolio/{address}")
    suspend fun getPortfolio(@Path("address") address: String): Response<ApiResponse<PortfolioResponse>>
}
