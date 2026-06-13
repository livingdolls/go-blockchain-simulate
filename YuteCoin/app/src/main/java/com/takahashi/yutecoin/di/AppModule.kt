package com.takahashi.yutecoin.di

import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.AuthRepository
import com.takahashi.yutecoin.ui.auth.LoginViewModel
import com.takahashi.yutecoin.ui.auth.RegisterViewModel
import org.koin.android.ext.koin.androidContext
import org.koin.core.module.dsl.viewModel
import org.koin.dsl.module

val appModule = module {
    single { AuthRepository() }
    single { SessionManager(androidContext()) }
    viewModel { LoginViewModel(get(), get()) }
    viewModel { RegisterViewModel(get(), get()) }
}
