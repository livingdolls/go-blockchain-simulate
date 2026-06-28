package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class PortfolioResponse(
    @SerializedName("address") val address: String,
    @SerializedName("yte_balance") val yteBalance: Double,
    @SerializedName("usd_balance") val usdBalance: Double,
    @SerializedName("yte_price") val ytePrice: Double,
    @SerializedName("total_value_usd") val totalValueUsd: Double,
    @SerializedName("total_deposited") val totalDeposited: Double,
    @SerializedName("total_withdrawn") val totalWithdrawn: Double,
    @SerializedName("total_traded") val totalTraded: Double,
    @SerializedName("realized_pnl") val realizedPnl: Double,
    @SerializedName("unrealized_pnl") val unrealizedPnl: Double,
    @SerializedName("pnl_percent") val pnlPercent: Double,
    @SerializedName("allocation") val allocation: Allocation
)

data class Allocation(
    @SerializedName("yte_percent") val ytePercent: Double,
    @SerializedName("usd_percent") val usdPercent: Double
)
