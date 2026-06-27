package com.takahashi.yutecoin.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.ui.unit.IntSize
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.local.ThemeManager
import com.takahashi.yutecoin.ui.auth.LoginScreen
import com.takahashi.yutecoin.ui.auth.RegisterScreen
import com.takahashi.yutecoin.ui.dashboard.DashboardScreen
import com.takahashi.yutecoin.ui.theme.RevealController

@Composable
fun AppNavHost(
    sessionManager: SessionManager,
    themeManager: ThemeManager,
    revealController: RevealController,
    rootSize: IntSize,
    navController: NavHostController = rememberNavController()
) {
    val startDestination = if (sessionManager.isLoggedIn()) {
        NavRoutes.Dashboard.route
    } else {
        NavRoutes.Login.route
    }

    NavHost(
        navController = navController,
        startDestination = startDestination
    ) {
        composable(NavRoutes.Login.route) {
            LoginScreen(
                onNavigateHome = {
                    navController.navigate(NavRoutes.Dashboard.route) {
                        popUpTo(NavRoutes.Login.route) { inclusive = true }
                    }
                },
                onNavigateToRegister = {
                    navController.navigate(NavRoutes.Register.route)
                }
            )
        }

        composable(NavRoutes.Register.route) {
            RegisterScreen(
                onNavigateHome = {
                    navController.navigate(NavRoutes.Dashboard.route) {
                        popUpTo(0) { inclusive = true }
                    }
                },
                onBack = {
                    navController.popBackStack()
                }
            )
        }

        composable(NavRoutes.Dashboard.route) {
            DashboardScreen(
                themeManager = themeManager,
                revealController = revealController,
                rootSize = rootSize,
                onLogout = {
                    sessionManager.clearAll()
                    navController.navigate(NavRoutes.Login.route) {
                        popUpTo(0) { inclusive = true }
                    }
                }
            )
        }
    }
}
