package com.takahashi.yutecoin.data.api

import android.util.Log
import com.google.gson.Gson
import com.takahashi.yutecoin.BuildConfig
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.BufferedReader
import java.io.InputStreamReader
import java.util.concurrent.TimeUnit

class SseClient {

    private val client = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(0, TimeUnit.MILLISECONDS)
        .build()

    private val gson = Gson()
    private val baseUrl = BuildConfig.SSE_BASE_URL

    fun <T> connect(path: String, type: Class<T>): Flow<T> = callbackFlow {
        val url = baseUrl + path
        Log.d("SseClient", "Connecting to SSE: $url")

        val request = Request.Builder()
            .url(url)
            .header("Accept", "text/event-stream")
            .header("Cache-Control", "no-cache")
            .build()

        val call = client.newCall(request)
        val response = withContext(Dispatchers.IO) { call.execute() }

        if (!response.isSuccessful) {
            Log.e("SseClient", "SSE connection failed: ${response.code}")
            close(Exception("SSE connection failed: ${response.code}"))
            return@callbackFlow
        }

        val body = response.body ?: run {
            Log.e("SseClient", "SSE response body is null")
            close(Exception("SSE response body is null"))
            return@callbackFlow
        }

        val reader = BufferedReader(InputStreamReader(body.byteStream()))
        val currentData = StringBuilder()

        awaitClose {
            Log.d("SseClient", "SSE connection closed, cancelling call")
            call.cancel()
            try { reader.close() } catch (_: Exception) {}
            try { response.close() } catch (_: Exception) {}
        }

        try {
            var line: String?
            while (isActive && reader.readLine().also { line = it } != null) {
                val l = line ?: break

                if (l.startsWith("data: ")) {
                    currentData.append(l.removePrefix("data: "))
                } else if (l.isEmpty() && currentData.isNotEmpty()) {
                    val json = currentData.toString().trim()
                    currentData.clear()

                    if (json.isNotEmpty()) {
                        try {
                            val parsed = gson.fromJson(json, type)
                            val result = trySend(parsed)
                            if (result.isFailure) {
                                Log.w("SseClient", "trySend failed, channel closed")
                                return@callbackFlow
                            }
                        } catch (e: Exception) {
                            Log.w("SseClient", "Failed to parse SSE data: $json", e)
                        }
                    }
                }
            }
            Log.d("SseClient", "SSE stream ended (eof)")
        } catch (e: Exception) {
            Log.e("SseClient", "SSE stream error", e)
        } finally {
            try { reader.close() } catch (_: Exception) {}
            try { response.close() } catch (_: Exception) {}
        }
    }
}
