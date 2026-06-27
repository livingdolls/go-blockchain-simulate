package com.takahashi.yutecoin.ui.theme

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.takahashi.yutecoin.data.local.ThemeManager
import kotlinx.coroutines.launch
import org.koin.compose.koinInject

@Composable
fun ThemeToggle(
    modifier: Modifier = Modifier
) {
    val themeManager = koinInject<ThemeManager>()
    val isDark by themeManager.isDarkMode.collectAsState(initial = false)
    val scope = rememberCoroutineScope()

    Text(
        text = if (isDark) "\u2600" else "\u263D",
        fontSize = 20.sp,
        modifier = modifier
            .padding(4.dp)
            .clickable {
                scope.launch {
                    themeManager.setDarkMode(!isDark)
                }
            }
    )
}
