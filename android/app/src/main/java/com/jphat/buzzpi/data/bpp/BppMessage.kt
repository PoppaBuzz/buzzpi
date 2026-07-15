package com.jphat.buzzpi.data.bpp

import org.json.JSONObject

object BppProtocol {
    const val VERSION = 1

    // Handshake methods
    const val METHOD_HANDSHAKE = "bpp.handshake"
    const val METHOD_AUTH_CHALLENGE = "bpp.auth.challenge"
    const val METHOD_AUTH_RESPONSE = "bpp.auth.response"
    const val METHOD_CAPABILITY_UPDATE = "bpp.capability.update"

    // Pairing methods
    const val METHOD_PAIR_INITIATE = "pair.initiate"
    const val METHOD_PAIR_VERIFY = "pair.verify"
    const val METHOD_PAIR_STATUS = "pair.status"
    const val METHOD_PAIR_UNPAIR = "pair.unpair"

    // Device methods
    const val METHOD_DEVICE_INFO = "device.info"
    const val METHOD_DEVICE_STATS = "device.stats"

    // Terminal methods
    const val METHOD_TERMINAL_OPEN = "terminal.open"
    const val METHOD_TERMINAL_INPUT = "terminal.input"
    const val METHOD_TERMINAL_RESIZE = "terminal.resize"
    const val METHOD_TERMINAL_CLOSE = "terminal.close"

    // Message types — matches Go agent envelope type constants
    const val TYPE_REQUEST = "request"
    const val TYPE_RESPONSE = "response"
    const val TYPE_EVENT = "event"
    const val TYPE_ERROR = "error"

    // Capability IDs — matches Go agent handshake capability constants
    const val CAP_TERMINAL = "terminal"
    const val CAP_SCREEN = "screen"
    const val CAP_FILE = "file"
    const val CAP_SYSTEM = "system"
    const val CAP_PLUGIN = "plugin"
    const val CAP_GPIO = "gpio"
    const val CAP_CAMERA = "camera"
    const val CAP_AUDIO = "audio"
    const val CAP_CLIPBOARD = "clipboard"
    const val CAP_NOTIFICATION = "notification"
    const val CAP_ENCRYPTION = "encryption"
    const val CAP_COMPRESSION = "compression"

    val DEFAULT_CAPABILITIES = listOf(
        CAP_TERMINAL, CAP_FILE, CAP_SYSTEM, CAP_PLUGIN
    )
}

/**
 * BPP Envelope — universal wrapper for all BPP messages.
 * Matches Go agent's Envelope struct (pkg/bpp/envelope.go) exactly.
 */
data class BppEnvelope(
    val version: Int = BppProtocol.VERSION,
    val id: String = "",
    val ts: String = "",
    val type: String,
    val method: String = "",
    val rid: String = "",
    val params: ByteArray? = null,
    val result: ByteArray? = null,
    val error: BppError? = null
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is BppEnvelope) return false
        return version == other.version &&
                id == other.id &&
                ts == other.ts &&
                type == other.type &&
                method == other.method &&
                rid == other.rid &&
                params.contentEquals(other.params ?: ByteArray(0)) &&
                result.contentEquals(other.result ?: ByteArray(0)) &&
                error == other.error
    }

    override fun hashCode(): Int {
        var result1 = version
        result1 = 31 * result1 + id.hashCode()
        result1 = 31 * result1 + ts.hashCode()
        result1 = 31 * result1 + type.hashCode()
        result1 = 31 * result1 + method.hashCode()
        result1 = 31 * result1 + rid.hashCode()
        result1 = 31 * result1 + (params?.contentHashCode() ?: 0)
        result1 = 31 * result1 + (result?.contentHashCode() ?: 0)
        result1 = 31 * result1 + (error?.hashCode() ?: 0)
        return result1
    }
}

data class BppError(
    val code: String,
    val message: String
)

/**
 * CapabilityOffer — sent by the connecting side to announce capabilities
 * and negotiate the protocol version.
 * Matches Go agent's CapabilityOffer struct (pkg/bpp/handshake.go).
 */
data class CapabilityOffer(
    val v: Int = BppProtocol.VERSION,
    val minV: Int = 1,
    val caps: List<String> = BppProtocol.DEFAULT_CAPABILITIES,
    val session: String = "",
    val deviceId: String = "",
    val clientName: String = "",
    val clientVersion: String = ""
)

/**
 * AuthChallenge — sent by agent when client has no valid session.
 * Matches Go agent's AuthChallenge struct (pkg/bpp/handshake.go).
 */
data class AuthChallenge(
    val type: String = "",
    val pin: String = "",
    val deviceId: String = "",
    val deviceName: String = "",
    val expiresAt: Long = 0L
)

/**
 * AuthResponse — sent by client to respond to an AuthChallenge.
 * Matches Go agent's AuthResponse struct (pkg/bpp/handshake.go).
 */
data class AuthResponse(
    val pin: String = "",
    val deviceId: String = "",
    val clientName: String = ""
)

/**
 * SessionEstablished — sent by agent when authentication succeeds.
 * Matches Go agent's SessionEstablished struct (pkg/bpp/handshake.go).
 */
data class SessionEstablished(
    val session: String = "",
    val deviceId: String = "",
    val deviceName: String = "",
    val caps: List<String> = emptyList(),
    val v: Int = BppProtocol.VERSION,
    val expiresAt: Long = 0L
)

data class InfoResponse(
    val deviceId: String = "",
    val friendlyName: String = "",
    val model: String = "",
    val runtimeVersion: String = "",
    val bppVersion: Int = 1,
    val uptimeSeconds: Long = 0L,
    val capabilities: List<String> = emptyList(),
    val platform: String = ""
)

data class StatsResponse(
    val cpu: CPUStats = CPUStats(),
    val memory: MemoryStats = MemoryStats(),
    val storage: List<DiskStats> = emptyList(),
    val network: NetworkStats = NetworkStats(),
    val uptimeSeconds: Long = 0L
)

data class CPUStats(
    val usagePercent: Double = 0.0,
    val temperatureCelsius: Double = 0.0,
    val frequencyMhz: Int = 0
)

data class MemoryStats(
    val totalMb: Long = 0L,
    val usedMb: Long = 0L,
    val availableMb: Long = 0L,
    val percent: Double = 0.0
)

data class DiskStats(
    val mount: String = "",
    val totalMb: Long = 0L,
    val usedMb: Long = 0L,
    val availableMb: Long = 0L,
    val percent: Double = 0.0
)

data class NetworkStats(
    val interfaces: List<InterfaceStats> = emptyList()
)

data class InterfaceStats(
    val name: String = "",
    val rxBytes: Long = 0L,
    val txBytes: Long = 0L
)

data class PairInitiateResult(
    val sessionId: String,
    val pin: String
)

data class PairVerifyResult(
    val sessionToken: String,
    val deviceId: String = ""
)
