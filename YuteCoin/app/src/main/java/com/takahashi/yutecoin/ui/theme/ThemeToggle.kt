package com.takahashi.yutecoin.ui.theme

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.layout.positionInRoot
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.takahashi.yutecoin.data.local.ThemeManager

@Composable
fun ThemeToggle(
    themeManager: ThemeManager,
    revealController: RevealController,
    rootSize: IntSize,
    modifier: Modifier = Modifier
) {
    val isDark by themeManager.isDarkMode.collectAsState()
    var position by remember { mutableStateOf(Offset.Zero) }

    Text(
        text = if (isDark) "\u2600" else "\u263D",
        fontSize = 20.sp,
        modifier = modifier
            .padding(4.dp)
            .onGloballyPositioned { coords ->
                val p = coords.positionInRoot()
                position = Offset(p.x + coords.size.width / 2f, p.y + coords.size.height / 2f)
            }
            .clickable {
                revealController.trigger(position, rootSize)
            }
    )
}
