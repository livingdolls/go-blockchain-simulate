package com.takahashi.yutecoin.di

import com.takahashi.yutecoin.data.api.SseClient
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.local.ThemeManager
import com.takahashi.yutecoin.data.repository.AuthRepository
import com.takahashi.yutecoin.data.repository.BalanceRepository
import com.takahashi.yutecoin.data.repository.MarketRepository
import com.takahashi.yutecoin.data.service.CandleStreamService
import com.takahashi.yutecoin.ui.auth.LoginViewModel
import com.takahashi.yutecoin.ui.auth.RegisterViewModel
import com.takahashi.yutecoin.ui.dashboard.HomeViewModel
import org.koin.android.ext.koin.androidApplication
import org.koin.android.ext.koin.androidContext
import org.koin.core.module.dsl.viewModel
import org.koin.dsl.module

val appModule = module {
    single { AuthRepository() }
    single { BalanceRepository() }
    single { MarketRepository() }
    single { SseClient() }
    single { CandleStreamService(get()) }
    single { SessionManager(androidContext()) }
    single { ThemeManager(androidContext()) }
    viewModel { LoginViewModel(androidApplication(), get(), get()) }
    viewModel { RegisterViewModel(androidApplication(), get(), get()) }
    viewModel { HomeViewModel(androidApplication(), get(), get(), get(), get()) }
}
