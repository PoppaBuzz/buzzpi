package com.jphat.buzzpi.ui.screens.pairing

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.domain.model.PairingResult
import com.jphat.buzzpi.domain.model.PairingSession
import com.jphat.buzzpi.domain.repository.DeviceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class PairingUiState(
    val stage: PairingStage = PairingStage.INITIATING,
    val session: PairingSession? = null,
    val result: PairingResult? = null,
    val error: String? = null
)

enum class PairingStage {
    INITIATING, WAITING_FOR_CONFIRMATION, VERIFYING, COMPLETED, FAILED
}

@HiltViewModel
class PairingViewModel
@Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val deviceRepository: DeviceRepository
) : ViewModel() {

    private val deviceId: String = savedStateHandle["deviceId"] ?: ""
    private val _uiState = MutableStateFlow(PairingUiState())
    val uiState: StateFlow<PairingUiState> = _uiState.asStateFlow()

    init {
        startPairing()
    }

    private fun startPairing() {
        viewModelScope.launch {
            try {
                val session = deviceRepository.initiatePairing(deviceId)
                _uiState.value = PairingUiState(
                    stage = PairingStage.WAITING_FOR_CONFIRMATION,
                    session = session
                )
                // Start polling for pairing status
                pollPairingStatus()
            } catch (e: Exception) {
                _uiState.value = PairingUiState(
                    stage = PairingStage.FAILED,
                    error = e.message ?: "Failed to initiate pairing"
                )
            }
        }
    }

    private suspend fun pollPairingStatus() {
        var attempts = 0
        val maxAttempts = 120 // 2 minutes at 1s intervals

        while (attempts < maxAttempts) {
            delay(1000)
            try {
                val result = deviceRepository.checkPairingStatus(deviceId)
                if (result != null) {
                    _uiState.value = PairingUiState(
                        stage = if (result.success) PairingStage.COMPLETED else PairingStage.FAILED,
                        result = result
                    )
                    return
                }
            } catch (_: Exception) {
                // Continue polling
            }
            attempts++
        }

        _uiState.value = PairingUiState(
            stage = PairingStage.FAILED,
            error = "Pairing timed out"
        )
    }

    fun retry() {
        _uiState.value = PairingUiState(stage = PairingStage.INITIATING)
        startPairing()
    }

    fun dismiss() {
        // Let navigation handle this
    }
}
