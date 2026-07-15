package com.jphat.buzzpi.domain.model

data class Session(
    val sessionToken: String,
    val deviceId: String,
    val clientName: String,
    val createdAt: Long = System.currentTimeMillis(),
    val expiresAt: Long = 0L
) {
    val isExpired: Boolean
        get() = expiresAt > 0 && System.currentTimeMillis() > expiresAt
}

data class PairingSession(
    val sessionId: String,
    val deviceId: String,
    val pin: String,
    val clientName: String
)

data class PairingResult(
    val success: Boolean,
    val sessionToken: String = "",
    val error: String? = null
)
