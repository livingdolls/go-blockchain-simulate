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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
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
    onOpenBlocks: () -> Unit = {},
    modifier: Modifier = Modifier
) {
    var showMenu by remember { mutableStateOf(false) }

    Column(modifier = modifier.fillMaxWidth(), horizontalAlignment = Alignment.End) {
        if (showMenu) {
            Column(
                modifier = Modifier
                    .padding(end = 24.dp, bottom = 4.dp)
                    .widthIn(max = 150.dp)
                    .clip(RoundedCornerShape(12.dp))
                    .background(MaterialTheme.colorScheme.surface.copy(alpha = 0.95f))
                    .border(0.5.dp, MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.3f), RoundedCornerShape(12.dp))
                    .padding(horizontal = 10.dp, vertical = 6.dp)
            ) {
                Row(verticalAlignment = Alignment.CenterVertically, modifier = Modifier.padding(vertical = 2.dp)) {
                    Text("Theme", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(Modifier.width(8.dp))
                    ThemeToggle(themeManager, revealController, rootSize)
                }
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.2f), thickness = 0.5.dp)
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier
                        .clickable { showMenu = false; onOpenBlocks() }
                        .padding(vertical = 4.dp)
                ) {
                    Text("Blocks", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.2f), thickness = 0.5.dp)
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier
                        .clickable { showMenu = false; onLogout() }
                        .padding(vertical = 4.dp)
                ) {
                    Text("Logout", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.error, fontWeight = FontWeight.SemiBold)
                }
            }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp)
                .clip(RoundedCornerShape(20.dp))
                .background(MaterialTheme.colorScheme.surface.copy(alpha = 0.75f))
                .border(0.5.dp, MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.3f), RoundedCornerShape(20.dp))
                .padding(horizontal = 4.dp, vertical = 4.dp),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically
        ) {
            BottomTab.entries.forEach { tab ->
                val sel = tab == currentTab
                val fg by animateColorAsState(if (sel) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f), label = "fg")
                val bg by animateColorAsState(if (sel) MaterialTheme.colorScheme.primary.copy(alpha = 0.9f) else Color.Transparent, label = "bg")
                Row(
                    modifier = Modifier.clip(RoundedCornerShape(18.dp)).background(bg).selectable(sel) { onTabSelected(tab) }.padding(horizontal = if (sel) 12.dp else 10.dp, vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(tab.unicodeIcon, fontSize = if (sel) 14.sp else 18.sp, color = fg)
                    if (sel) { Spacer(Modifier.width(5.dp)); Text(tab.label, style = MaterialTheme.typography.labelSmall, fontWeight = FontWeight.Bold, color = fg, fontSize = 11.sp) }
                }
            }

            Row(
                modifier = Modifier.clip(RoundedCornerShape(18.dp)).selectable(showMenu) { showMenu = !showMenu }.padding(horizontal = 10.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("\u22EE", fontSize = 18.sp, color = if (showMenu) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
            }
        }
    }
}
