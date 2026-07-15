package com.jphat.buzzpi.ui.screens.qrpairing

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.data.bpp.BppClient
import com.jphat.buzzpi.data.bpp.HandshakeHandler
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.json.JSONObject
import javax.inject.Inject

data class QrPairingUiState(
    val isManualMode: Boolean = false,
    val manualCode: String = "",
    val isPairing: Boolean = false,
    val pairingComplete: Boolean = false,
    val deviceId: String = "",
    val deviceName: String = "",
    val error: String? = null
)

@HiltViewModel
class QrPairingViewModel
@Inject constructor(
    private val bppClient: BppClient,
    private val handshakeHandler: HandshakeHandler
) : ViewModel() {

    private val _uiState = MutableStateFlow(QrPairingUiState())
    val uiState: StateFlow<QrPairingUiState> = _uiState.asStateFlow()

    fun onQrCodeScanned(code: String) {
        if (_uiState.value.isPairing || _uiState.value.pairingComplete) return
        _uiState.value = _uiState.value.copy(manualCode = code)
        connectToDevice(code)
    }

    fun toggleManualMode() {
        _uiState.value = _uiState.value.copy(
            isManualMode = !_uiState.value.isManualMode,
            error = null
        )
    }

    fun updateManualCode(code: String) {
        _uiState.value = _uiState.value.copy(manualCode = code)
    }

    fun submitManualCode() {
        val code = _uiState.value.manualCode.trim()
        if (code.isNotEmpty()) {
            connectToDevice(code)
        }
    }

    private fun connectToDevice(deviceAddress: String) {
        _uiState.value = _uiState.value.copy(isPairing = true, error = null)
        viewModelScope.launch {
            try {
                val url = if (deviceAddress.startsWith("ws")) {
                    deviceAddress
                } else {
                    "ws://$deviceAddress:8443/ws"
                }

                val handshakeResult = handshakeHandler.performHandshake(url)
                val challenge = handshakeResult.getOrThrow().challenge

                _uiState.value = _uiState.value.copy(
                    isPairing = false,
                    pairingComplete = true,
                    deviceId = challenge.deviceId,
                    deviceName = challenge.deviceName
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isPairing = false,
                    error = e.message ?: "Pairing failed"
                )
            }
        }
    }

    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
}
