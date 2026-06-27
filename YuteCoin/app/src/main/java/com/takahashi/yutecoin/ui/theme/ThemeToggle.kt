package com.takahashi.yutecoin.ui.theme

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.takahashi.yutecoin.data.local.ThemeManager

@Composable
fun ThemeToggle(
    themeManager: ThemeManager,
    modifier: Modifier = Modifier
) {
    val isDark by themeManager.isDarkMode.collectAsState()

    Text(
        text = if (isDark) "\u2600" else "\u263D",
        fontSize = 20.sp,
        modifier = modifier
            .padding(4.dp)
            .clickable { themeManager.toggle() }
    )
}
