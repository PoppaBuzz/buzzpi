package com.jphat.buzzpi.domain.repository

import com.jphat.buzzpi.domain.model.Device
import com.jphat.buzzpi.domain.model.Dimensions
import com.jphat.buzzpi.domain.model.PairingResult
import com.jphat.buzzpi.domain.model.PairingSession
import com.jphat.buzzpi.domain.model.Session
import com.jphat.buzzpi.domain.model.TerminalInput
import com.jphat.buzzpi.domain.model.TerminalState
import com.jphat.buzzpi.data.bpp.StatsResponse
import kotlinx.coroutines.flow.Flow

interface DeviceRepository {
    val discoveredDevices: Flow<List<Device>>
    val pairedDevices: Flow<List<Device>>
    val terminalState: Flow<TerminalState>

    suspend fun startDiscovery()
    suspend fun stopDiscovery()
    suspend fun pair(device: Device): PairingResult
    suspend fun unpair(deviceId: String)
    suspend fun getDeviceInfo(deviceId: String): Device?
    suspend fun initiatePairing(deviceId: String): PairingSession
    suspend fun checkPairingStatus(deviceId: String): PairingResult?
    suspend fun getDevice(deviceId: String): Device?
    suspend fun connectToDevice(device: Device, session: Session)
    suspend fun disconnectFromDevice(deviceId: String)
    suspend fun disconnect(deviceId: String) = disconnectFromDevice(deviceId)

    // Live device info and stats
    suspend fun fetchDeviceInfo(deviceId: String): Device?
    suspend fun fetchDeviceStats(deviceId: String): StatsResponse?

    // Terminal
    suspend fun connectTerminal(deviceId: String)
    suspend fun disconnectTerminal(deviceId: String)
    suspend fun sendTerminalInput(input: TerminalInput)
    suspend fun clearTerminal(deviceId: String)
    suspend fun resizeTerminal(deviceId: String, dimensions: Dimensions)
}
