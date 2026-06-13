package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class RegisterRequest(
    @SerializedName("username") val username: String,
    @SerializedName("address") val address: String,
    @SerializedName("public_key") val publicKey: String
)

data class RegisterResponse(
    @SerializedName("id") val id: Int,
    @SerializedName("username") val username: String,
    @SerializedName("address") val address: String,
    @SerializedName("public_key") val publicKey: String,
    @SerializedName("token") val token: String? = null
)

data class ChallengeResponse(
    @SerializedName("challenge") val challenge: String
)

data class VerifyRequest(
    @SerializedName("address") val address: String,
    @SerializedName("signature") val signature: String,
    @SerializedName("nonce") val nonce: String,
    @SerializedName("username") val username: String
)

data class ApiResponse<T>(
    @SerializedName("success") val success: Boolean,
    @SerializedName("data") val data: T?,
    @SerializedName("message") val message: String?,
    @SerializedName("errors") val errors: Map<String, List<String>>?
)
