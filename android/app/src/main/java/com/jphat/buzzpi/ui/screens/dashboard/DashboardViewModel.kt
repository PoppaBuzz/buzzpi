package com.jphat.buzzpi.ui.screens.dashboard

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.data.bpp.CPUStats
import com.jphat.buzzpi.data.bpp.DiskStats
import com.jphat.buzzpi.data.bpp.MemoryStats
import com.jphat.buzzpi.data.bpp.NetworkStats
import com.jphat.buzzpi.data.bpp.StatsResponse
import com.jphat.buzzpi.domain.model.Device
import com.jphat.buzzpi.domain.repository.DeviceRepository
import com.jphat.buzzpi.domain.repository.SessionRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class DashboardUiState(
    val device: Device? = null,
    val isLoading: Boolean = true,
    val status: DeviceStatus = DeviceStatus.UNKNOWN,
    val error: String? = null,
    val cpu: CPUStats? = null,
    val memory: MemoryStats? = null,
    val storage: List<DiskStats>? = null,
    val network: NetworkStats? = null
)

enum class DeviceStatus {
    ONLINE, OFFLINE, UNKNOWN
}

@HiltViewModel
class DashboardViewModel
@Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val deviceRepository: DeviceRepository,
    private val sessionRepository: SessionRepository
) : ViewModel() {

    private val deviceId: String = savedStateHandle["deviceId"] ?: ""
    private val _uiState = MutableStateFlow(DashboardUiState())
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    init {
        loadDevice()
    }

    private fun loadDevice() {
        viewModelScope.launch {
            try {
                val device = deviceRepository.getDevice(deviceId)
                val state = DashboardUiState(
                    device = device,
                    isLoading = false,
                    status = if (device?.isOnline == true) DeviceStatus.ONLINE else DeviceStatus.OFFLINE
                )
                _uiState.value = state

                if (device != null) {
                    val session = sessionRepository.getSession(deviceId)
                    if (session != null) {
                        try {
                            deviceRepository.connectToDevice(device, session)
                            _uiState.value = _uiState.value.copy(status = DeviceStatus.ONLINE)
                            deviceRepository.fetchDeviceInfo(deviceId)
                            val stats = deviceRepository.fetchDeviceStats(deviceId)
                            if (stats != null) {
                                _uiState.value = _uiState.value.copy(
                                    cpu = stats.cpu,
                                    memory = stats.memory,
                                    storage = stats.storage,
                                    network = stats.network
                                )
                            }
                        } catch (_: Exception) {
                            _uiState.value = _uiState.value.copy(
                                status = DeviceStatus.OFFLINE,
                                error = "Connection failed \u2014 tap refresh to retry"
                            )
                        }
                    }
                }
            } catch (e: Exception) {
                _uiState.value = DashboardUiState(
                    isLoading = false,
                    error = e.message ?: "Failed to load device"
                )
            }
        }
    }

    fun disconnect() {
        viewModelScope.launch {
            try {
                deviceRepository.disconnect(deviceId)
                sessionRepository.clearAllSessions()
            } catch (_: Exception) { }
        }
    }

    fun refresh() {
        _uiState.value = _uiState.value.copy(
            isLoading = true, error = null,
            cpu = null, memory = null, storage = null, network = null
        )
        loadDevice()
    }
}
