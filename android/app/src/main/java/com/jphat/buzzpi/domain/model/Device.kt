package com.jphat.buzzpi.domain.model

data class Device(
    val deviceId: String,
    val friendlyName: String,
    val platform: String = "",
    val runtimeVersion: String = "",
    val capabilities: List<String> = emptyList(),
    val isOnline: Boolean = true,
    val transport: Transport = Transport.LAN,
    val ipAddress: String = "",
    val port: Int = 0
)

enum class Transport {
    LAN,
    RELAY,
    P2P
}
