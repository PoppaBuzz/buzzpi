package com.jphat.buzzpi.ui.navigation

sealed class Routes(val route: String) {
    data object Discovery : Routes("discovery")

    data object Pairing : Routes("pairing/{deviceId}") {
        fun createRoute(deviceId: String) = "pairing/$deviceId"
    }

    data object Dashboard : Routes("dashboard/{deviceId}") {
        fun createRoute(deviceId: String) = "dashboard/$deviceId"
    }

    data object Terminal : Routes("terminal/{deviceId}") {
        fun createRoute(deviceId: String) = "terminal/$deviceId"
    }

    data object ScreenViewer : Routes("screenviewer/{deviceId}") {
        fun createRoute(deviceId: String) = "screenviewer/$deviceId"
    }

    data object FileManager : Routes("filemanager/{deviceId}") {
        fun createRoute(deviceId: String) = "filemanager/$deviceId"
    }

    data object QrPairing : Routes("qrpairing")

    data object Settings : Routes("settings")
}
