package com.takahashi.yutecoin.data.dto

import com.google.gson.Gson
import com.google.gson.JsonSyntaxException
import com.google.gson.annotations.SerializedName
import com.google.gson.reflect.TypeToken

data class RegisterRequest(
    @SerializedName("username") val username: String,
    @SerializedName("address") val address: String,
    @SerializedName("public_key") val publicKey: String
)

data class RegisterResponse(
    @SerializedName("username") val username: String,
    @SerializedName("address") val address: String,
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
    @SerializedName("code") val code: Int? = null,
    @SerializedName("data") val data: T?,
    @SerializedName("error") val error: String? = null,
    @SerializedName("error_code") val errorCode: String? = null,
    @SerializedName("field") val field: String? = null,
    @SerializedName("details") val details: List<FieldError>? = null
)

data class FieldError(
    @SerializedName("field") val field: String,
    @SerializedName("message") val message: String
) {
    companion object {
        val gson = Gson()
        val listType = object : TypeToken<List<FieldError>>() {}.type

        fun listFromJson(json: String): List<FieldError> {
            return try {
                gson.fromJson(json, listType)
            } catch (e: JsonSyntaxException) {
                emptyList()
            }
        }
    }
}
