package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class BalanceResponse(
    @SerializedName("name") val name: String,
    @SerializedName("address") val address: String,
    @SerializedName("yte_balance") val yteBalance: Double,
    @SerializedName("usd_balance") val usdBalance: Double
)
