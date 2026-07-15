package com.jphat.buzzpi.data.bpp

import org.json.JSONArray
import org.json.JSONObject

class HandshakeHandler(
    private val bppClient: BppClient
) {
    private var deviceId: String = ""
    private var deviceName: String = ""

    /**
     * @return true if session was accepted (connection ready), false if expired
     */
    suspend fun reconnectWithSession(
        url: String,
        sessionToken: String,
        clientName: String = "BuzzPi Android"
    ): Boolean {
        return try {
            bppClient.connect(url)

            val offer = JSONObject().apply {
                put("v", BppProtocol.VERSION)
                put("min_v", 1)
                put("caps", JSONArray(BppProtocol.DEFAULT_CAPABILITIES))
                put("session", sessionToken)
                put("client_name", clientName)
                put("client_version", "1.0.0")
            }

            val response = bppClient.sendRequest(BppProtocol.METHOD_HANDSHAKE, offer)
            if (response.error != null) return false

            val resultData = response.result?.let { JSONObject(it.decodeToString()) }
                ?: return false

            val newSession = resultData.optString("session", "")
            if (newSession.isNotEmpty()) {
                deviceId = resultData.optString("device_id", "")
                deviceName = resultData.optString("device_name", "")
                return true
            }
            false
        } catch (_: Exception) {
            false
        }
    }

    suspend fun performHandshake(
        url: String,
        clientName: String = "BuzzPi Android"
    ): Result<HandshakeResult> {
        return try {
            bppClient.connect(url)

            // Send CapabilityOffer in Go agent wire format
            val offer = JSONObject().apply {
                put("v", BppProtocol.VERSION)
                put("min_v", 1)
                put("caps", JSONArray(BppProtocol.DEFAULT_CAPABILITIES))
                put("client_name", clientName)
                put("client_version", "1.0.0")
            }

            val challengeResponse = bppClient.sendRequest(BppProtocol.METHOD_HANDSHAKE, offer)
            if (challengeResponse.error != null) {
                return Result.failure(Exception("Handshake error: ${challengeResponse.error.message}"))
            }

            val challengeData = challengeResponse.result?.let { JSONObject(it.decodeToString()) }
                ?: return Result.failure(Exception("Empty handshake response"))

            // Parse AuthChallenge — matches Go AuthChallenge struct
            val challenge = AuthChallenge(
                type = challengeData.optString("type", ""),
                pin = challengeData.optString("pin", ""),
                deviceId = challengeData.optString("device_id", ""),
                deviceName = challengeData.optString("device_name", ""),
                expiresAt = challengeData.optLong("expires_at", 0L)
            )
            this.deviceId = challenge.deviceId
            this.deviceName = challenge.deviceName

            Result.success(
                HandshakeResult(
                    challenge = challenge,
                    needsPairing = true
                )
            )
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Sends an AuthResponse with the PIN to complete authentication.
     * Uses Go agent's "bpp.auth.response" method and AuthResponse struct.
     */
    suspend fun authenticateWithPin(
        pin: String,
        clientName: String = "BuzzPi Android"
    ): Result<SessionEstablished> {
        return try {
            val authParams = JSONObject().apply {
                put("pin", pin)
                put("device_id", deviceId)
                put("client_name", clientName)
            }

            val response = bppClient.sendRequest(BppProtocol.METHOD_AUTH_RESPONSE, authParams)
            if (response.error != null) {
                return Result.failure(Exception("Auth failed: ${response.error.message}"))
            }

            val resultData = response.result?.let { JSONObject(it.decodeToString()) }
                ?: return Result.failure(Exception("Empty auth response"))

            // Parse SessionEstablished — matches Go SessionEstablished struct
            val session = SessionEstablished(
                session = resultData.optString("session", ""),
                deviceId = resultData.optString("device_id", ""),
                deviceName = resultData.optString("device_name", ""),
                caps = resultData.optJSONArray("caps")?.let { arr ->
                    (0 until arr.length()).map { arr.optString(it) }
                } ?: emptyList(),
                v = resultData.optInt("v", BppProtocol.VERSION),
                expiresAt = resultData.optLong("expires_at", 0L)
            )

            if (session.session.isEmpty()) {
                return Result.failure(Exception("Empty session token"))
            }
            Result.success(session)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun pairInitiate(
        clientName: String = "BuzzPi Android"
    ): Result<PairInitiateResult> {
        return try {
            val params = JSONObject().apply {
                put("device_id", deviceId)
                put("client_name", clientName)
                put("supported_auths", JSONArray(listOf("pin")))
            }

            val response = bppClient.sendRequest(BppProtocol.METHOD_PAIR_INITIATE, params)
            if (response.error != null) {
                return Result.failure(Exception("Pair initiate failed: ${response.error.message}"))
            }

            val data = response.result?.let { JSONObject(it.decodeToString()) }
                ?: return Result.failure(Exception("Empty pair.initiate response"))

            Result.success(
                PairInitiateResult(
                    sessionId = data.optString("session_id", ""),
                    pin = data.optString("pin", "")
                )
            )
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun pairVerify(
        sessionId: String,
        pin: String,
        clientName: String = "BuzzPi Android"
    ): Result<PairVerifyResult> {
        return try {
            val params = JSONObject().apply {
                put("session_id", sessionId)
                put("pin", pin)
                put("client_name", clientName)
            }

            val response = bppClient.sendRequest(BppProtocol.METHOD_PAIR_VERIFY, params)
            if (response.error != null) {
                return Result.failure(Exception("Pair verify failed: ${response.error.message}"))
            }

            val data = response.result?.let { JSONObject(it.decodeToString()) }
                ?: return Result.failure(Exception("Empty pair.verify response"))

            // PairVerifyResult from Go agent: session_token, expires_at, device_id
            Result.success(
                PairVerifyResult(
                    sessionToken = data.optString("session_token", ""),
                    deviceId = data.optString("device_id", "")
                )
            )
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

data class HandshakeResult(
    val challenge: AuthChallenge,
    val needsPairing: Boolean
)
