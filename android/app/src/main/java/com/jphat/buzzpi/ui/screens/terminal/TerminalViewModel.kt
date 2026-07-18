package com.jphat.buzzpi.ui.screens.terminal

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.domain.model.Dimensions
import com.jphat.buzzpi.domain.model.TerminalInput
import com.jphat.buzzpi.domain.model.TerminalState
import com.jphat.buzzpi.domain.repository.DeviceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class TerminalUiState(
    val terminalState: TerminalState = TerminalState(),
    val isConnected: Boolean = false,
    val isConnecting: Boolean = true,
    val inputBuffer: String = "",
    val cursorPosition: Int = 0,
    val error: String? = null
)

@HiltViewModel
class TerminalViewModel
@Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val deviceRepository: DeviceRepository
) : ViewModel() {

    private val deviceId: String = savedStateHandle["deviceId"] ?: ""
    private val _uiState = MutableStateFlow(TerminalUiState())
    val uiState: StateFlow<TerminalUiState> = _uiState.asStateFlow()
    private var terminalJob: Job? = null
    private val history = mutableListOf<String>()
    private var historyIndex = -1

    init {
        connectTerminal()
    }

    private fun connectTerminal() {
        _uiState.value = _uiState.value.copy(isConnecting = true, error = null)
        terminalJob = viewModelScope.launch {
            try {
                deviceRepository.connectTerminal(deviceId)
                deviceRepository.terminalState.collect { state ->
                    _uiState.value = _uiState.value.copy(
                        terminalState = state,
                        isConnected = true,
                        isConnecting = false,
                        error = null
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isConnecting = false,
                    isConnected = false,
                    error = e.message ?: "Failed to connect to terminal"
                )
            }
        }
    }

    fun sendInput(text: String) {
        if (text.isBlank() || !_uiState.value.isConnected) return

        // Add to history
        history.add(text)
        historyIndex = history.size

        viewModelScope.launch {
            try {
                val input = TerminalInput(
                    data = text.encodeToByteArray()
                )
                deviceRepository.sendTerminalInput(input)
                _uiState.value = _uiState.value.copy(inputBuffer = "", cursorPosition = 0)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Failed to send input"
                )
            }
        }
    }

    fun sendSpecialKey(key: ByteArray) {
        if (!_uiState.value.isConnected) return
        viewModelScope.launch {
            try {
                deviceRepository.sendTerminalInput(TerminalInput(data = key))
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Failed to send key"
                )
            }
        }
    }

    fun updateInputBuffer(text: String) {
        _uiState.value = _uiState.value.copy(
            inputBuffer = text,
            cursorPosition = text.length
        )
    }

    fun navigateHistory(direction: HistoryDirection) {
        when (direction) {
            HistoryDirection.UP -> {
                if (historyIndex > 0) {
                    historyIndex--
                    val previous = history.getOrElse(historyIndex) { "" }
                    _uiState.value = _uiState.value.copy(
                        inputBuffer = previous,
                        cursorPosition = previous.length
                    )
                }
            }
            HistoryDirection.DOWN -> {
                if (historyIndex < history.size - 1) {
                    historyIndex++
                    val next = history.getOrElse(historyIndex) { "" }
                    _uiState.value = _uiState.value.copy(
                        inputBuffer = next,
                        cursorPosition = next.length
                    )
                } else {
                    historyIndex = history.size
                    _uiState.value = _uiState.value.copy(
                        inputBuffer = "",
                        cursorPosition = 0
                    )
                }
            }
        }
    }

    fun clearTerminal() {
        viewModelScope.launch {
            try {
                deviceRepository.clearTerminal(deviceId)
                _uiState.value = _uiState.value.copy(
                    terminalState = TerminalState()
                )
            } catch (_: Exception) { }
        }
    }

    fun resizeTerminal(rows: Int, cols: Int) {
        viewModelScope.launch {
            try {
                deviceRepository.resizeTerminal(deviceId, Dimensions(rows = rows, cols = cols))
            } catch (_: Exception) { }
        }
    }

    fun reconnect() {
        terminalJob?.cancel()
        connectTerminal()
    }

    override fun onCleared() {
        super.onCleared()
        terminalJob?.cancel()
        viewModelScope.launch {
            deviceRepository.disconnectTerminal(deviceId)
        }
    }
}

enum class HistoryDirection {
    UP, DOWN
}
