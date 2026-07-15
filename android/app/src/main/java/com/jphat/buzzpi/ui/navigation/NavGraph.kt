package com.jphat.buzzpi.ui.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.jphat.buzzpi.ui.screens.dashboard.DashboardScreen
import com.jphat.buzzpi.ui.screens.discovery.DiscoveryScreen
import com.jphat.buzzpi.ui.screens.filemanager.FileManagerScreen
import com.jphat.buzzpi.ui.screens.pairing.PairingScreen
import com.jphat.buzzpi.ui.screens.qrpairing.QrPairingScreen
import com.jphat.buzzpi.ui.screens.screenviewer.ScreenViewerScreen
import com.jphat.buzzpi.ui.screens.settings.SettingsScreen
import com.jphat.buzzpi.ui.screens.terminal.TerminalScreen

@Composable
fun BuzzPiNavGraph(navController: NavHostController) {
    NavHost(
        navController = navController,
        startDestination = Routes.Discovery.route
    ) {
        composable(Routes.Discovery.route) {
            DiscoveryScreen(
                onNavigateToPairing = { deviceId ->
                    navController.navigate(Routes.Pairing.createRoute(deviceId))
                },
                onNavigateToDashboard = { deviceId ->
                    navController.navigate(Routes.Dashboard.createRoute(deviceId))
                },
                onNavigateToTerminal = { deviceId ->
                    navController.navigate(Routes.Terminal.createRoute(deviceId))
                },
                onNavigateToSettings = {
                    navController.navigate(Routes.Settings.route)
                }
            )
        }

        composable(
            route = Routes.Pairing.route,
            arguments = listOf(
                navArgument("deviceId") { type = NavType.StringType }
            )
        ) {
            PairingScreen(
                onNavigateBack = { navController.popBackStack() },
                onPairingComplete = { deviceId ->
                    navController.navigate(Routes.Terminal.createRoute(deviceId)) {
                        popUpTo(Routes.Discovery.route)
                    }
                }
            )
        }

        composable(
            route = Routes.Dashboard.route,
            arguments = listOf(
                navArgument("deviceId") { type = NavType.StringType }
            )
        ) {
            DashboardScreen(
                onNavigateToTerminal = { deviceId ->
                    navController.navigate(Routes.Terminal.createRoute(deviceId))
                },
                onNavigateToScreenViewer = { deviceId ->
                    navController.navigate(Routes.ScreenViewer.createRoute(deviceId))
                },
                onNavigateToFileManager = { deviceId ->
                    navController.navigate(Routes.FileManager.createRoute(deviceId))
                },
                onNavigateBack = { navController.popBackStack() }
            )
        }

        composable(
            route = Routes.Terminal.route,
            arguments = listOf(
                navArgument("deviceId") { type = NavType.StringType }
            )
        ) {
            TerminalScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }

        composable(
            route = Routes.ScreenViewer.route,
            arguments = listOf(
                navArgument("deviceId") { type = NavType.StringType }
            )
        ) {
            ScreenViewerScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }

        composable(
            route = Routes.FileManager.route,
            arguments = listOf(
                navArgument("deviceId") { type = NavType.StringType }
            )
        ) {
            FileManagerScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }

        composable(Routes.QrPairing.route) {
            QrPairingScreen(
                onNavigateBack = { navController.popBackStack() },
                onPairingComplete = { deviceId ->
                    navController.navigate(Routes.Terminal.createRoute(deviceId)) {
                        popUpTo(Routes.Discovery.route)
                    }
                }
            )
        }

        composable(Routes.Settings.route) {
            SettingsScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }
    }
}
