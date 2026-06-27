package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class MarketResponse(
    @SerializedName("id") val id: Int,
    @SerializedName("price") val price: Double,
    @SerializedName("liquidity") val liquidity: Double,
    @SerializedName("last_block") val lastBlock: Long,
    @SerializedName("updated_at") val updatedAt: String
)

data class CandleResponse(
    @SerializedName("id") val id: Long,
    @SerializedName("interval_type") val intervalType: String,
    @SerializedName("start_time") val startTime: Long,
    @SerializedName("open_price") val openPrice: Double,
    @SerializedName("high_price") val highPrice: Double,
    @SerializedName("low_price") val lowPrice: Double,
    @SerializedName("close_price") val closePrice: Double,
    @SerializedName("volume") val volume: Double
)
