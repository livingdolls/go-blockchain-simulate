package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.WalletResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Path
import retrofit2.http.Query

interface WalletApi {

    @GET("wallet/{address}")
    suspend fun getTransactions(
        @Path("address") address: String,
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): Response<WalletResponse>
}
