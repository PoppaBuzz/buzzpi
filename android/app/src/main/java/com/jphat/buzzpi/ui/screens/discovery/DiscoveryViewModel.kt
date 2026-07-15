package com.jphat.buzzpi.ui.screens.discovery

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.domain.model.Device
import com.jphat.buzzpi.domain.repository.DeviceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class DiscoveryUiState(
    val isScanning: Boolean = true,
    val devices: List<Device> = emptyList(),
    val pairedDevices: List<Device> = emptyList(),
    val error: String? = null
)

@HiltViewModel
class DiscoveryViewModel
@Inject constructor(
    private val deviceRepository: DeviceRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(DiscoveryUiState())
    val uiState: StateFlow<DiscoveryUiState> = _uiState.asStateFlow()

    private val _selectedDevice = MutableStateFlow<Device?>(null)
    val selectedDevice: StateFlow<Device?> = _selectedDevice.asStateFlow()

    init {
        viewModelScope.launch {
            deviceRepository.discoveredDevices.collect { devices ->
                _uiState.update { copy(devices = devices, isScanning = false) }
            }
        }
        viewModelScope.launch {
            deviceRepository.pairedDevices.collect { paired ->
                _uiState.update { copy(pairedDevices = paired) }
            }
        }
        startDiscovery()
    }

    fun startDiscovery() {
        _uiState.update { copy(isScanning = true, error = null) }
        viewModelScope.launch {
            try {
                deviceRepository.startDiscovery()
            } catch (e: Exception) {
                _uiState.update { copy(error = e.message, isScanning = false) }
            }
        }
    }

    fun refreshDiscovery() {
        startDiscovery()
    }

    fun selectDevice(device: Device) {
        _selectedDevice.value = device
    }

    fun clearSelection() {
        _selectedDevice.value = null
    }

    fun pairWithDevice(device: Device) {
        viewModelScope.launch {
            try {
                _uiState.update { copy(error = null) }
                val result = deviceRepository.pair(device)
                if (!result.success) {
                    _uiState.update { copy(error = result.error ?: "Pairing failed") }
                }
            } catch (e: Exception) {
                _uiState.update { copy(error = e.message) }
            }
        }
    }

    override fun onCleared() {
        super.onCleared()
        viewModelScope.launch {
            deviceRepository.stopDiscovery()
        }
    }
}

private fun MutableStateFlow<DiscoveryUiState>.update(transform: DiscoveryUiState.() -> DiscoveryUiState) {
    value = transform(value)
}
