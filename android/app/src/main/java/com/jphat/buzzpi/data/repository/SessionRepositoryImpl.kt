package com.jphat.buzzpi.data.repository

import android.content.Context
import android.content.SharedPreferences
import com.jphat.buzzpi.domain.model.Session
import com.jphat.buzzpi.domain.repository.SessionRepository
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import org.json.JSONObject
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SessionRepositoryImpl
@Inject constructor(
    @param:ApplicationContext private val context: Context
) : SessionRepository {

    private val prefs: SharedPreferences =
        context.getSharedPreferences("buzzpi_sessions", Context.MODE_PRIVATE)

    private val _activeSession = MutableStateFlow<Session?>(null)
    override val activeSession: Flow<Session?> = _activeSession.asStateFlow()

    override suspend fun saveSession(session: Session) {
        val json = JSONObject().apply {
            put("session_token", session.sessionToken)
            put("device_id", session.deviceId)
            put("client_name", session.clientName)
            put("created_at", session.createdAt)
            put("expires_at", session.expiresAt)
        }
        prefs.edit().putString("session_${session.deviceId}", json.toString()).apply()
        _activeSession.value = session
    }

    override suspend fun getSession(deviceId: String): Session? {
        val jsonStr = prefs.getString("session_$deviceId", null) ?: return null
        return try {
            val json = JSONObject(jsonStr)
            Session(
                sessionToken = json.optString("session_token", ""),
                deviceId = json.optString("device_id", ""),
                clientName = json.optString("client_name", ""),
                createdAt = json.optLong("created_at", 0L),
                expiresAt = json.optLong("expires_at", 0L)
            )
        } catch (e: Exception) {
            null
        }
    }

    override suspend fun deleteSession(deviceId: String) {
        prefs.edit().remove("session_$deviceId").apply()
        if (_activeSession.value?.deviceId == deviceId) {
            _activeSession.value = null
        }
    }

    override suspend fun clearAllSessions() {
        prefs.edit().clear().apply()
        _activeSession.value = null
    }

    fun restoreActiveSession(): Session? {
        val allEntries = prefs.all
        for ((key, value) in allEntries) {
            if (key.startsWith("session_")) {
                try {
                    val json = JSONObject(value as? String ?: continue)
                    val session = Session(
                        sessionToken = json.optString("session_token", ""),
                        deviceId = json.optString("device_id", ""),
                        clientName = json.optString("client_name", ""),
                        createdAt = json.optLong("created_at", 0L),
                        expiresAt = json.optLong("expires_at", 0L)
                    )
                    if (!session.isExpired) {
                        _activeSession.value = session
                        return session
                    }
                } catch (_: Exception) { }
            }
        }
        return null
    }
}
