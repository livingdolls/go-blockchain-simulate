package com.takahashi.yutecoin.ui.navigation

import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.takahashi.yutecoin.data.local.ThemeManager
import com.takahashi.yutecoin.ui.theme.RevealController
import com.takahashi.yutecoin.ui.theme.ThemeToggle

enum class BottomTab(
    val label: String,
    val unicodeIcon: String
) {
    HOME("Home", "\u2302"),
    TRADE("Trade", "\u2194"),
    STAKING("Earn", "\u25C8"),
    PORTFOLIO("Port", "\u25B3")
}

@Composable
fun FloatingBottomNavBar(
    currentTab: BottomTab,
    onTabSelected: (BottomTab) -> Unit,
    themeManager: ThemeManager,
    revealController: RevealController,
    rootSize: IntSize,
    onLogout: () -> Unit,
    modifier: Modifier = Modifier
) {
    var showMenu by remember { mutableStateOf(false) }

    Column(
        modifier = modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        if (showMenu) {
            Surface(
                modifier = Modifier
                    .padding(horizontal = 32.dp),
                shape = RoundedCornerShape(16.dp),
                color = MaterialTheme.colorScheme.surface.copy(alpha = 0.92f),
                shadowElevation = 8.dp,
                tonalElevation = 4.dp
            ) {
                Column(
                    modifier = Modifier.padding(12.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            text = "Theme",
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                        ThemeToggle(
                            themeManager = themeManager,
                            revealController = revealController,
                            rootSize = rootSize,
                            modifier = Modifier.padding(4.dp)
                        )
                    }

                    Spacer(modifier = Modifier.height(4.dp))

                    Surface(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable {
                                showMenu = false
                                onLogout()
                            },
                        shape = RoundedCornerShape(10.dp),
                        color = Color.Transparent
                    ) {
                        Text(
                            text = "Logout",
                            style = MaterialTheme.typography.labelMedium,
                            color = MaterialTheme.colorScheme.error,
                            fontWeight = FontWeight.SemiBold,
                            modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(4.dp))
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp)
                .clip(RoundedCornerShape(20.dp))
                .background(
                    MaterialTheme.colorScheme.surface.copy(alpha = 0.75f)
                )
                .border(
                    width = 0.5.dp,
                    color = MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.3f),
                    shape = RoundedCornerShape(20.dp)
                )
                .padding(horizontal = 4.dp, vertical = 4.dp),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically
        ) {
            BottomTab.entries.forEach { tab ->
                val isSelected = tab == currentTab
                val fgColor by animateColorAsState(
                    targetValue = if (isSelected)
                        MaterialTheme.colorScheme.onPrimary
                    else
                        MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                    label = "tabFg"
                )
                val bgColor by animateColorAsState(
                    targetValue = if (isSelected)
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.9f)
                    else
                        Color.Transparent,
                    label = "tabBg"
                )

                Row(
                    modifier = Modifier
                        .clip(RoundedCornerShape(18.dp))
                        .background(bgColor)
                        .selectable(
                            selected = isSelected,
                            onClick = { onTabSelected(tab) }
                        )
                        .padding(
                            horizontal = if (isSelected) 12.dp else 10.dp,
                            vertical = 8.dp
                        ),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = tab.unicodeIcon,
                        fontSize = if (isSelected) 14.sp else 18.sp,
                        color = fgColor
                    )
                    if (isSelected) {
                        Spacer(modifier = Modifier.width(5.dp))
                        Text(
                            text = tab.label,
                            style = MaterialTheme.typography.labelSmall,
                            fontWeight = FontWeight.Bold,
                            color = fgColor,
                            fontSize = 11.sp
                        )
                    }
                }
            }

            val menuFg = if (showMenu)
                MaterialTheme.colorScheme.primary
            else
                MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)

            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(18.dp))
                    .selectable(
                        selected = showMenu,
                        onClick = { showMenu = !showMenu }
                    )
                    .padding(horizontal = 10.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "\u22EE",
                    fontSize = 18.sp,
                    color = menuFg
                )
            }
        }
    }
}
