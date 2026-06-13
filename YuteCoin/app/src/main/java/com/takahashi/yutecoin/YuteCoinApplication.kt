package com.takahashi.yutecoin

import android.app.Application
import android.util.Log
import com.takahashi.yutecoin.di.appModule
import org.koin.android.ext.koin.androidContext
import org.koin.core.context.startKoin

class YuteCoinApplication : Application() {
    override fun onCreate() {
        super.onCreate()

        // Global crash logger — delegate to default handler
        val defaultHandler = Thread.getDefaultUncaughtExceptionHandler()
        Thread.setDefaultUncaughtExceptionHandler { thread, e ->
            Log.e("YuteCoin", "FATAL CRASH: ${e.javaClass.simpleName}: ${e.message}", e)
            defaultHandler?.uncaughtException(thread, e)
        }

        startKoin {
            androidContext(this@YuteCoinApplication)
            modules(appModule)
        }

        // Load BIP39 wordlist from assets immediately (in background to not block startup)
        Thread {
            try {
                com.takahashi.yutecoin.crypto.Bip39.loadFromAssets(this@YuteCoinApplication)
            } catch (e: Exception) {
                Log.e("YuteCoin", "Failed to load BIP39 wordlist", e)
            }
        }.start()
    }
}
