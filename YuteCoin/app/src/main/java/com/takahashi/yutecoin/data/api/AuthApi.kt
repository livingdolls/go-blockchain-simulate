package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.data.dto.ApiResponse
import com.takahashi.yutecoin.data.dto.ChallengeResponse
import com.takahashi.yutecoin.data.dto.RegisterRequest
import com.takahashi.yutecoin.data.dto.RegisterResponse
import com.takahashi.yutecoin.data.dto.VerifyRequest
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path

interface AuthApi {

    @POST("register")
    suspend fun register(@Body request: RegisterRequest): Response<ApiResponse<RegisterResponse>>

    @POST("challenge/{address}")
    suspend fun getChallenge(@Path("address") address: String): Response<ApiResponse<ChallengeResponse>>

    @POST("challenge/verify")
    suspend fun verifyChallenge(@Body request: VerifyRequest): Response<ApiResponse<Unit>>

    @GET("profile")
    suspend fun getProfile(): Response<ApiResponse<RegisterResponse>>
}
