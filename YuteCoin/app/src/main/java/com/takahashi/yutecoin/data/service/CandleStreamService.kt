package com.takahashi.yutecoin.data.service

import android.util.Log
import com.takahashi.yutecoin.data.api.SseClient
import com.takahashi.yutecoin.data.dto.CandleResponse
import kotlinx.coroutines.flow.Flow

class CandleStreamService(
    private val sseClient: SseClient
) {

    fun streamCandles(interval: String = "1m"): Flow<CandleResponse> {
        Log.d("CandleStream", "Starting candle stream for interval: $interval")
        return sseClient.connect("candles?interval=$interval", CandleResponse::class.java)
    }
}
