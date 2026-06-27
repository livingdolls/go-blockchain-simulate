package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class NonceResponse(
    @SerializedName("nonce") val nonce: String
)

data class SendRequest(
    @SerializedName("from_address") val fromAddress: String,
    @SerializedName("to_address") val toAddress: String,
    @SerializedName("amount") val amount: Double,
    @SerializedName("nonce") val nonce: String,
    @SerializedName("signature") val signature: String
)

data class SendResponse(
    @SerializedName("message") val message: String
)
