package com.jphat.buzzpi.data.repository

import com.jphat.buzzpi.data.bpp.BppClient
import com.jphat.buzzpi.data.bpp.HandshakeHandler
import com.jphat.buzzpi.data.bpp.StatsResponse
import com.jphat.buzzpi.data.discovery.MdnsDiscovery
import com.jphat.buzzpi.domain.model.Device
import com.jphat.buzzpi.domain.model.Dimensions
import com.jphat.buzzpi.domain.model.PairingResult
import com.jphat.buzzpi.domain.model.PairingSession
import com.jphat.buzzpi.domain.model.Session
import com.jphat.buzzpi.domain.model.TerminalInput
import com.jphat.buzzpi.domain.model.TerminalLine
import com.jphat.buzzpi.domain.model.TerminalState
import com.jphat.buzzpi.domain.repository.DeviceRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.json.JSONObject
import javax.inject.Inject
import javax.inject.Singleton
import dagger.hilt.android.qualifiers.ApplicationContext

@Singleton
class DeviceRepositoryImpl
@Inject constructor(
    private val mdnsDiscovery: MdnsDiscovery,
    private val bppClient: BppClient,
    private val handshakeHandler: HandshakeHandler,
    private val sessionRepository: SessionRepositoryImpl,
    @param:dagger.hilt.android.qualifiers.ApplicationContext private val context: android.content.Context
) : DeviceRepository {

    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private var terminalOutputJob: Job? = null

    private val _pairedDevices = MutableStateFlow<List<Device>>(emptyList())
    override val pairedDevices: Flow<List<Device>> = _pairedDevices.asStateFlow()
    override val discoveredDevices: Flow<List<Device>> = mdnsDiscovery.discoveredDevices

    private val _terminalState = MutableStateFlow(TerminalState())
    override val terminalState: Flow<TerminalState> = _terminalState.asStateFlow()

    init {
        loadPairedDevices()
    }

    private fun loadPairedDevices() {
        val prefs = context.getSharedPreferences("buzzpi_paired_devices", android.content.Context.MODE_PRIVATE)
        val json = prefs.getString("paired_devices", null) ?: return
        try {
            val arr = org.json.JSONArray(json)
            val devices = (0 until arr.length()).mapNotNull { i ->
                val obj = arr.getJSONObject(i)
                Device(
                    deviceId = obj.optString("device_id", ""),
                    friendlyName = obj.optString("friendly_name", ""),
                    ipAddress = obj.optString("ip_address", ""),
                    port = obj.optInt("port", 8420),
                    runtimeVersion = obj.optString("runtime_version", ""),
                    platform = obj.optString("platform", ""),
                    isOnline = true
                )
            }.filter { it.deviceId.isNotEmpty() }
            _pairedDevices.value = devices
        } catch (_: Exception) { }
    }

    private fun savePairedDevices() {
        val prefs = context.getSharedPreferences("buzzpi_paired_devices", android.content.Context.MODE_PRIVATE)
        val arr = org.json.JSONArray()
        _pairedDevices.value.forEach { device ->
            arr.put(org.json.JSONObject().apply {
                put("device_id", device.deviceId)
                put("friendly_name", device.friendlyName)
                put("ip_address", device.ipAddress)
                put("port", device.port)
                put("runtime_version", device.runtimeVersion)
                put("platform", device.platform)
            })
        }
        prefs.edit().putString("paired_devices", arr.toString()).apply()
    }

    override suspend fun startDiscovery() {
        mdnsDiscovery.startDiscovery()
    }

    override suspend fun stopDiscovery() {
        mdnsDiscovery.stopDiscovery()
    }

    override suspend fun pair(device: Device): PairingResult {
        return try {
            val wsUrl = "ws://${device.ipAddress}:${device.port}/ws"
            val result = handshakeHandler.performHandshake(wsUrl, "BuzzPi Android")
            if (result.isFailure) {
                return PairingResult(false, error = result.exceptionOrNull()?.message)
            }

            val handshakeResult = result.getOrNull()
                ?: return PairingResult(false, error = "Empty handshake result")

            val authResult = handshakeHandler.authenticateWithPin(
                pin = handshakeResult.challenge.pin,
                clientName = "BuzzPi Android"
            )
            if (authResult.isFailure) {
                return PairingResult(false, error = authResult.exceptionOrNull()?.message)
            }

            val session = authResult.getOrNull()
                ?: return PairingResult(false, error = "Empty session")

            sessionRepository.saveSession(
                Session(
                    sessionToken = session.session,
                    deviceId = device.deviceId,
                    clientName = "BuzzPi Android",
                    expiresAt = session.expiresAt
                )
            )

            val current = _pairedDevices.value.toMutableList()
            if (current.none { it.deviceId == device.deviceId }) {
                current.add(device)
                _pairedDevices.value = current
                savePairedDevices()
            }

            PairingResult(true, sessionToken = session.session)
        } catch (e: Exception) {
            PairingResult(false, error = e.message ?: "Unknown error")
        }
    }

    override suspend fun initiatePairing(deviceId: String): PairingSession {
        val result = handshakeHandler.pairInitiate("BuzzPi Android")
        if (result.isFailure) {
            throw Exception(result.exceptionOrNull()?.message ?: "Pair initiation failed")
        }

        val data = result.getOrNull() ?: throw Exception("Empty pair init result")
        return PairingSession(
            sessionId = data.sessionId,
            deviceId = deviceId,
            pin = data.pin,
            clientName = "BuzzPi Android"
        )
    }

    override suspend fun checkPairingStatus(deviceId: String): PairingResult? {
        return try {
            val session = sessionRepository.getSession(deviceId)
            if (session != null && !session.isExpired) {
                PairingResult(true, sessionToken = session.sessionToken)
            } else null
        } catch (_: Exception) {
            null
        }
    }

    override suspend fun getDevice(deviceId: String): Device? {
        return _pairedDevices.value.find { it.deviceId == deviceId }
    }

    override suspend fun getDeviceInfo(deviceId: String): Device? {
        return getDevice(deviceId)
    }

    override suspend fun unpair(deviceId: String) {
        try {
            bppClient.sendMessage("pair.unpair", JSONObject().apply { put("device_id", deviceId) })
        } catch (_: Exception) { }

        val current = _pairedDevices.value.toMutableList()
        current.removeAll { it.deviceId == deviceId }
        _pairedDevices.value = current
        savePairedDevices()
        sessionRepository.deleteSession(deviceId)
    }

    override suspend fun fetchDeviceInfo(deviceId: String): Device? {
        val response = bppClient.sendRequest("device.info", JSONObject())
        if (response.error != null) return null

        val data = response.result?.let { JSONObject(it.decodeToString()) } ?: return null
        val device = _pairedDevices.value.find { it.deviceId == deviceId } ?: return null

        val updated = device.copy(
            friendlyName = data.optString("name", device.friendlyName),
            runtimeVersion = data.optString("bpp_version", device.runtimeVersion),
            platform = data.optString("os", device.platform)
        )

        val current = _pairedDevices.value.toMutableList()
        val idx = current.indexOfFirst { it.deviceId == deviceId }
        if (idx >= 0) current[idx] = updated
        _pairedDevices.value = current

        return updated
    }

    override suspend fun fetchDeviceStats(deviceId: String): StatsResponse? {
        val response = bppClient.sendRequest("device.stats", JSONObject())
        if (response.error != null) return null

        val data = response.result?.let { JSONObject(it.decodeToString()) } ?: return null

        return StatsResponse(
            cpu = com.jphat.buzzpi.data.bpp.CPUStats(
                usagePercent = data.optJSONObject("cpu")?.optDouble("usage_percent", 0.0) ?: 0.0,
                temperatureCelsius = data.optJSONObject("cpu")?.optDouble("temperature_celsius", 0.0) ?: 0.0,
                frequencyMhz = data.optJSONObject("cpu")?.optInt("frequency_mhz", 0) ?: 0
            ),
            memory = com.jphat.buzzpi.data.bpp.MemoryStats(
                totalMb = data.optJSONObject("memory")?.optLong("total_mb", 0L) ?: 0L,
                usedMb = data.optJSONObject("memory")?.optLong("used_mb", 0L) ?: 0L,
                availableMb = data.optJSONObject("memory")?.optLong("available_mb", 0L) ?: 0L,
                percent = data.optJSONObject("memory")?.optDouble("percent", 0.0) ?: 0.0
            ),
            storage = data.optJSONArray("storage")?.let { arr ->
                (0 until arr.length()).map { idx ->
                    val disk = arr.getJSONObject(idx)
                    com.jphat.buzzpi.data.bpp.DiskStats(
                        mount = disk.optString("mount", ""),
                        totalMb = disk.optLong("total_mb", 0L),
                        usedMb = disk.optLong("used_mb", 0L),
                        availableMb = disk.optLong("available_mb", 0L),
                        percent = disk.optDouble("percent", 0.0)
                    )
                }
            } ?: emptyList(),
            network = com.jphat.buzzpi.data.bpp.NetworkStats(
                interfaces = data.optJSONObject("network")?.optJSONArray("interfaces")?.let { arr ->
                    (0 until arr.length()).map { idx ->
                        val iface = arr.getJSONObject(idx)
                        com.jphat.buzzpi.data.bpp.InterfaceStats(
                            name = iface.optString("name", ""),
                            rxBytes = iface.optLong("rx_bytes", 0L),
                            txBytes = iface.optLong("tx_bytes", 0L)
                        )
                    }
                } ?: emptyList()
            ),
            uptimeSeconds = data.optLong("uptime_seconds", 0L)
        )
    }

    override suspend fun connectToDevice(device: Device, session: Session) {
        val wsUrl = "ws://${device.ipAddress}:${device.port}/ws"
        val accepted = handshakeHandler.reconnectWithSession(wsUrl, session.sessionToken)
        if (!accepted) {
            throw Exception("Session expired — re-pairing required")
        }
    }

    override suspend fun disconnectFromDevice(deviceId: String) {
        bppClient.disconnect()
    }

    override suspend fun ensureConnected(deviceId: String) {
        if (bppClient.isConnected()) return

        val device = _pairedDevices.value.find { it.deviceId == deviceId }
            ?: throw Exception("Device $deviceId not found in paired devices")

        val session = sessionRepository.getSession(deviceId)
            ?: throw Exception("No session found for $deviceId")

        if (session.isExpired) throw Exception("Session expired — re-pairing required")

        val wsUrl = "ws://${device.ipAddress}:${device.port}/ws"
        val accepted = handshakeHandler.reconnectWithSession(wsUrl, session.sessionToken)
        if (!accepted) throw Exception("Session rejected by device — re-pairing required")
    }

    override suspend fun connectTerminal(deviceId: String) {
        _terminalState.value = _terminalState.value.copy(isConnected = false)

        val device = _pairedDevices.value.find { it.deviceId == deviceId }
        val session = device?.let { sessionRepository.getSession(deviceId) }

        if (device == null || session == null) {
            _terminalState.value = _terminalState.value.copy(
                error = "Device not paired or session expired. Please re-pair."
            )
            return
        }

        val wsUrl = "ws://${device.ipAddress}:${device.port}/ws"
        val accepted = handshakeHandler.reconnectWithSession(wsUrl, session.sessionToken)
        if (!accepted) {
            _terminalState.value = _terminalState.value.copy(error = "Session expired — re-pair required")
            return
        }

        // Use sendRequest to get session_id back from Go agent
        val params = JSONObject().apply {
            put("device_id", deviceId)
            put("cols", _terminalState.value.dimensions.cols)
            put("rows", _terminalState.value.dimensions.rows)
        }
        val response = bppClient.sendRequest("terminal.open", params)

        if (response.error != null) {
            _terminalState.value = _terminalState.value.copy(
                error = response.error!!.message
            )
            return
        }

        val resultJson = response.result?.let { JSONObject(it.decodeToString()) }
        val sessionId = resultJson?.optString("session_id", "") ?: ""

        _terminalState.value = _terminalState.value.copy(
            isConnected = true,
            sessionId = sessionId
        )

        terminalOutputJob?.cancel()
        terminalOutputJob = scope.launch {
            bppClient.messages.collect { envelope ->
                when (envelope.method) {
                    "terminal.output" -> {
                        val params = envelope.params?.let { JSONObject(it.decodeToString()) }
                        val text = params?.optString("data", "") ?: ""
                        if (text.isNotEmpty()) {
                            val line = TerminalLine(text = text)
                            val current = _terminalState.value.lines.toMutableList()
                            current.add(line)
                            _terminalState.value = _terminalState.value.copy(lines = current)
                        }
                    }
                    "terminal.closed" -> {
                        _terminalState.value = _terminalState.value.copy(isConnected = false)
                        terminalOutputJob = null
                    }
                }
            }
        }
    }

    override suspend fun disconnectTerminal(deviceId: String) {
        terminalOutputJob?.cancel()
        terminalOutputJob = null
        try {
            bppClient.sendMessage("terminal.close", JSONObject().apply {
                put("device_id", deviceId)
            })
        } catch (_: Exception) { }
        _terminalState.value = _terminalState.value.copy(isConnected = false)
    }

    override suspend fun sendTerminalInput(input: TerminalInput) {
        val sessionId = _terminalState.value.sessionId
        val params = JSONObject().apply {
            put("data", input.data.decodeToString())
            if (sessionId.isNotEmpty()) put("session_id", sessionId)
        }
        val response = bppClient.sendRequest("terminal.input", params)

        if (response.error != null) {
            val current = _terminalState.value.lines.toMutableList()
            current.add(TerminalLine(text = "Error: ${response.error!!.message}"))
            _terminalState.value = _terminalState.value.copy(lines = current)
        }
    }

    override suspend fun clearTerminal(deviceId: String) {
        _terminalState.value = TerminalState()
    }

    override suspend fun resizeTerminal(deviceId: String, dimensions: Dimensions) {
        val sessionId = _terminalState.value.sessionId
        bppClient.sendMessage("terminal.resize", JSONObject().apply {
            if (sessionId.isNotEmpty()) put("session_id", sessionId)
            put("cols", dimensions.cols)
            put("rows", dimensions.rows)
        })
        _terminalState.value = _terminalState.value.copy(dimensions = dimensions)
    }
}
