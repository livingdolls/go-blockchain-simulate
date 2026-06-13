package com.takahashi.yutecoin

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.ui.navigation.AppNavHost
import com.takahashi.yutecoin.ui.theme.YuteCoinTheme
import org.koin.android.ext.android.inject

class MainActivity : ComponentActivity() {

    private val sessionManager: SessionManager by inject()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            YuteCoinTheme {
                AppNavHost(sessionManager = sessionManager)
            }
        }
    }
}
