package com.takahashi.yutecoin.data.api

import com.takahashi.yutecoin.BuildConfig
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

object RetrofitClient {

    val cookieJar = PersistentCookieJar()

    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = if (BuildConfig.DEBUG) HttpLoggingInterceptor.Level.BODY
        else HttpLoggingInterceptor.Level.NONE
    }

    val okHttpClient: OkHttpClient = OkHttpClient.Builder()
        .cookieJar(cookieJar)
        .addInterceptor(loggingInterceptor)
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .build()

    private val retrofit: Retrofit = Retrofit.Builder()
        .baseUrl(BuildConfig.API_BASE_URL)
        .client(okHttpClient)
        .addConverterFactory(GsonConverterFactory.create())
        .build()

    val authApi: AuthApi = retrofit.create(AuthApi::class.java)
    val balanceApi: BalanceApi = retrofit.create(BalanceApi::class.java)
    val marketApi: MarketApi = retrofit.create(MarketApi::class.java)
    val transactionApi: TransactionApi = retrofit.create(TransactionApi::class.java)
    val walletApi: WalletApi = retrofit.create(WalletApi::class.java)
    val stakingApi: StakingApi = retrofit.create(StakingApi::class.java)
    val portfolioApi: PortfolioApi = retrofit.create(PortfolioApi::class.java)

    fun clearSession() {
        cookieJar.clear()
    }
}
