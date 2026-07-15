package com.jphat.buzzpi.domain.repository

import com.jphat.buzzpi.domain.model.Session
import kotlinx.coroutines.flow.Flow

interface SessionRepository {
    val activeSession: Flow<Session?>
    suspend fun saveSession(session: Session)
    suspend fun getSession(deviceId: String): Session?
    suspend fun deleteSession(deviceId: String)
    suspend fun clearAllSessions()
}
