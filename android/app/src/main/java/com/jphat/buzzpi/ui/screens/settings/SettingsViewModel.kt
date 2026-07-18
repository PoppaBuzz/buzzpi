package com.jphat.buzzpi.ui.screens.settings

import android.content.Context
import android.content.SharedPreferences
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.data.bpp.BppClient
import com.jphat.buzzpi.domain.repository.DeviceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import org.json.JSONObject
import javax.inject.Inject

data class SettingsUiState(
    val deviceName: String = "",
    val deviceId: String = "",
    val platform: String = "",
    val version: String = "",
    val relayServer: String = "",
    val terminalFontSize: Int = 14,
    val terminalFontFamily: String = "monospace",
    val isSaving: Boolean = false,
    val isRenaming: Boolean = false,
    val saveSuccess: Boolean = false,
    val error: String? = null
)

@HiltViewModel
class SettingsViewModel
@Inject constructor(
    private val deviceRepository: DeviceRepository,
    private val bppClient: BppClient,
    @ApplicationContext context: Context
) : ViewModel() {

    private val prefs: SharedPreferences =
        context.getSharedPreferences("buzzpi_settings", Context.MODE_PRIVATE)

    private val _uiState = MutableStateFlow(SettingsUiState())
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    init {
        loadSettings()
    }

    private fun loadSettings() {
        viewModelScope.launch {
            try {
                val devices = deviceRepository.pairedDevices.first()
                val device = devices.firstOrNull()

                _uiState.value = _uiState.value.copy(
                    deviceName = device?.friendlyName ?: "",
                    deviceId = device?.deviceId ?: "",
                    platform = device?.platform ?: "",
                    version = device?.runtimeVersion ?: "",
                    relayServer = prefs.getString("relay_server", "") ?: "",
                    terminalFontSize = prefs.getInt("terminal_font_size", 14),
                    terminalFontFamily = prefs.getString("terminal_font_family", "monospace") ?: "monospace"
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Failed to load settings"
                )
            }
        }
    }

    fun updateDeviceName(name: String) {
        _uiState.value = _uiState.value.copy(deviceName = name)
    }

    fun renameDevice(newName: String) {
        if (newName.isBlank()) {
            _uiState.value = _uiState.value.copy(error = "Name must not be empty")
            return
        }
        _uiState.value = _uiState.value.copy(isRenaming = true, error = null)
        viewModelScope.launch {
            try {
                deviceRepository.ensureConnected(_uiState.value.deviceId)
                val params = org.json.JSONObject().apply { put("name", newName) }
                val response = bppClient.sendRequest("device.rename", params)
                if (response.error != null) {
                    _uiState.value = _uiState.value.copy(
                        isRenaming = false,
                        error = response.error.message ?: "Rename failed"
                    )
                    return@launch
                }
                val result = response.result?.let { org.json.JSONObject(it.decodeToString()) }
                val savedName = result?.optString("name", newName) ?: newName
                _uiState.value = _uiState.value.copy(
                    deviceName = savedName,
                    isRenaming = false,
                    saveSuccess = true
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isRenaming = false,
                    error = e.message ?: "Rename failed"
                )
            }
        }
    }

    fun updateRelayServer(url: String) {
        _uiState.value = _uiState.value.copy(relayServer = url)
    }

    fun updateTerminalFontSize(size: Int) {
        _uiState.value = _uiState.value.copy(terminalFontSize = size.coerceIn(10, 24))
    }

    fun updateTerminalFontFamily(family: String) {
        _uiState.value = _uiState.value.copy(terminalFontFamily = family)
    }

    fun saveSettings() {
        _uiState.value = _uiState.value.copy(isSaving = true, error = null)
        viewModelScope.launch {
            try {
                prefs.edit()
                    .putString("relay_server", _uiState.value.relayServer)
                    .putInt("terminal_font_size", _uiState.value.terminalFontSize)
                    .putString("terminal_font_family", _uiState.value.terminalFontFamily)
                    .apply()

                _uiState.value = _uiState.value.copy(
                    isSaving = false,
                    saveSuccess = true
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isSaving = false,
                    error = e.message ?: "Failed to save settings"
                )
            }
        }
    }

    fun clearSaveSuccess() {
        _uiState.value = _uiState.value.copy(saveSuccess = false)
    }

    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
}
