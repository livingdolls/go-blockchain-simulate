package com.takahashi.yutecoin.data.api

import android.util.Log
import okhttp3.Cookie
import okhttp3.CookieJar
import okhttp3.HttpUrl

/**
 * In-memory persistent CookieJar that stores cookies and replays them on subsequent requests.
 * For full persistence across app restarts, extend with DataStore serialization.
 */
class PersistentCookieJar : CookieJar {

    private val cookies = mutableMapOf<String, MutableList<Cookie>>()

    override fun saveFromResponse(url: HttpUrl, cookies: List<Cookie>) {
        val key = cookieKey(url)
        val existing = this.cookies.getOrPut(key) { mutableListOf() }

        // Update or add new cookies
        for (cookie in cookies) {
            existing.removeAll { it.name == cookie.name }
            existing.add(cookie)
            Log.d("CookieJar", "Saved cookie: ${cookie.name} = ${cookie.value.take(20)}...")
        }
    }

    override fun loadForRequest(url: HttpUrl): List<Cookie> {
        val key = cookieKey(url)
        val result = cookies[key]?.filter { !it.expiresAt.let { exp -> exp > 0 && exp < System.currentTimeMillis() / 1000 } }
        return result ?: emptyList()
    }

    fun clear() {
        cookies.clear()
        Log.d("CookieJar", "All cookies cleared")
    }

    fun hasAuthCookie(): Boolean {
        return cookies.values.flatten().any { it.name == "auth_token" || it.name == "admin_token" }
    }

    private fun cookieKey(url: HttpUrl): String {
        return "${url.host}:${url.port}"
    }
}
