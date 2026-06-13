package com.takahashi.yutecoin.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.ui.auth.LoginScreen
import com.takahashi.yutecoin.ui.auth.RegisterScreen
import com.takahashi.yutecoin.ui.dashboard.HomeScreen
import org.koin.androidx.compose.koinViewModel

@Composable
fun AppNavHost(
    sessionManager: SessionManager,
    navController: NavHostController = rememberNavController()
) {
    val startDestination = if (sessionManager.isLoggedIn()) {
        NavRoutes.Home.route
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
                    navController.navigate(NavRoutes.Home.route) {
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
                    navController.navigate(NavRoutes.Home.route) {
                        popUpTo(0) { inclusive = true }
                    }
                },
                onBack = {
                    navController.popBackStack()
                }
            )
        }

        composable(NavRoutes.Home.route) {
            HomeScreen(
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
