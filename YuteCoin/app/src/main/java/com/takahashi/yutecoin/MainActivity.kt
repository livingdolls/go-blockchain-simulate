package com.takahashi.yutecoin

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.local.ThemeManager
import com.takahashi.yutecoin.ui.navigation.AppNavHost
import com.takahashi.yutecoin.ui.theme.YuteCoinTheme
import org.koin.android.ext.android.inject

class MainActivity : ComponentActivity() {

    private val sessionManager: SessionManager by inject()
    private val themeManager: ThemeManager by inject()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            val isDarkMode by themeManager.isDarkMode.collectAsState(initial = false)

            YuteCoinTheme(darkTheme = isDarkMode) {
                AppNavHost(sessionManager = sessionManager)
            }
        }
    }
}
