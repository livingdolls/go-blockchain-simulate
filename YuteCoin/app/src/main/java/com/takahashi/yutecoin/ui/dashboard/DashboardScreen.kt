package com.takahashi.yutecoin.ui.dashboard

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.IntSize
import com.takahashi.yutecoin.data.local.ThemeManager
import com.takahashi.yutecoin.ui.navigation.BottomTab
import com.takahashi.yutecoin.ui.navigation.FloatingBottomNavBar
import com.takahashi.yutecoin.ui.theme.RevealController

@Composable
fun DashboardScreen(
    themeManager: ThemeManager,
    revealController: RevealController,
    rootSize: IntSize,
    onLogout: () -> Unit
) {
    var currentTab by remember { mutableStateOf(BottomTab.HOME) }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        bottomBar = {
            FloatingBottomNavBar(
                currentTab = currentTab,
                onTabSelected = { currentTab = it },
                themeManager = themeManager,
                revealController = revealController,
                rootSize = rootSize,
                onLogout = onLogout
            )
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            when (currentTab) {
                BottomTab.HOME -> HomeScreen()
                BottomTab.TRADE -> TradeScreen()
                BottomTab.STAKING -> StakingScreen()
                BottomTab.PORTFOLIO -> PortfolioScreen()
            }
        }
    }
}
