package com.takahashi.yutecoin

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.unit.IntSize
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.local.ThemeManager
import com.takahashi.yutecoin.ui.navigation.AppNavHost
import com.takahashi.yutecoin.ui.theme.RevealController
import com.takahashi.yutecoin.ui.theme.ThemeRevealBox
import com.takahashi.yutecoin.ui.theme.YuteCoinTheme
import com.takahashi.yutecoin.ui.theme.rememberRevealController
import org.koin.android.ext.android.inject

class MainActivity : ComponentActivity() {

    private val sessionManager: SessionManager by inject()
    private val themeManager: ThemeManager by inject()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            val revealController = rememberRevealController()
            var isDarkMode by remember { mutableStateOf(themeManager.isDarkModeInternal()) }
            var rootSize by remember { mutableStateOf(IntSize.Zero) }

            ThemeRevealBox(
                controller = revealController,
                onMidpoint = {
                    isDarkMode = !isDarkMode
                    themeManager.setDarkMode(isDarkMode)
                }
            ) {
                YuteCoinTheme(darkTheme = isDarkMode) {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .onSizeChanged { rootSize = it }
                    ) {
                        AppNavHost(
                            sessionManager = sessionManager,
                            themeManager = themeManager,
                            revealController = revealController,
                            rootSize = rootSize,
                            isDarkMode = isDarkMode
                        )
                    }
                }
            }
        }
    }
}
