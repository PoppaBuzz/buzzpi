package com.jphat.buzzpi.ui.screens.screenviewer

import android.graphics.BitmapFactory
import android.util.Base64
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.data.bpp.BppClient
import com.jphat.buzzpi.data.bpp.ConnectionState
import com.jphat.buzzpi.domain.repository.DeviceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.json.JSONObject
import javax.inject.Inject

data class ScreenViewerUiState(
    val frameBitmap: android.graphics.Bitmap? = null,
    val frameWidth: Int = 0,
    val frameHeight: Int = 0,
    val frameCount: Int = 0,
    val isStreaming: Boolean = false,
    val fps: Int = 1,
    val quality: Int = 70,
    val isConnected: Boolean = false,
    val isConnecting: Boolean = true,
    val error: String? = null
)

@HiltViewModel
class ScreenViewerViewModel
@Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val bppClient: BppClient,
    private val deviceRepository: DeviceRepository
) : ViewModel() {

    private val deviceId: String = savedStateHandle["deviceId"] ?: ""
    private val _uiState = MutableStateFlow(ScreenViewerUiState())
    val uiState: StateFlow<ScreenViewerUiState> = _uiState.asStateFlow()
    private var streamJob: Job? = null

    init {
        connectAndCapture()
    }

    private fun connectAndCapture() {
        _uiState.value = _uiState.value.copy(isConnecting = true, error = null)
        viewModelScope.launch {
            try {
                deviceRepository.ensureConnected(deviceId)
                _uiState.value = _uiState.value.copy(
                    isConnected = true,
                    isConnecting = false
                )
                captureScreenshot()
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isConnecting = false,
                    isConnected = false,
                    error = e.message ?: "Failed to connect"
                )
            }
        }
    }

    fun captureScreenshot() {
        viewModelScope.launch {
            try {
                val params = JSONObject().apply {
                    put("width", 1280)
                    put("height", 720)
                    put("quality", _uiState.value.quality)
                }
                val response = bppClient.sendRequest("screen.capture", params)

                if (response.error != null) {
                    val errorMsg = response.error!!.message
                    _uiState.value = _uiState.value.copy(
                        error = if (errorMsg.contains("capture") || errorMsg.contains("screen") || errorMsg.contains("display"))
                            "Screen capture unavailable. This Pi may not have a display connected."
                        else errorMsg
                    )
                    return@launch
                }

                val resultJson = response.result?.let { org.json.JSONObject(it.decodeToString()) }
                val data = resultJson?.optString("data", "") ?: ""
                if (data.isNotEmpty()) {
                    val bytes = Base64.decode(data, Base64.DEFAULT)
                    val bitmap = BitmapFactory.decodeByteArray(bytes, 0, bytes.size)
                    if (bitmap != null) {
                        _uiState.value = _uiState.value.copy(
                            frameBitmap = bitmap,
                            frameWidth = bitmap.width,
                            frameHeight = bitmap.height,
                            frameCount = _uiState.value.frameCount + 1
                        )
                        if (!_uiState.value.isStreaming) {
                            startStream()
                        }
                    }
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = "Screen capture failed: ${e.message}. This Pi may not have a display connected."
                )
            }
        }
    }

    fun startStream() {
        if (_uiState.value.isStreaming) return
        _uiState.value = _uiState.value.copy(isStreaming = true, error = null)

        streamJob = viewModelScope.launch {
            while (_uiState.value.isStreaming) {
                try {
                    val params = JSONObject().apply {
                        put("width", 1280)
                        put("height", 720)
                        put("quality", _uiState.value.quality)
                    }
                    val response = bppClient.sendRequest("screen.capture", params)

                    if (response.error != null) {
                        _uiState.value = _uiState.value.copy(
                            error = "Stream error: ${response.error!!.message}",
                            isStreaming = false
                        )
                        break
                    }

                    val resultJson = response.result?.let { org.json.JSONObject(it.decodeToString()) }
                    val data = resultJson?.optString("data", "") ?: ""
                    if (data.isNotEmpty()) {
                        val bytes = Base64.decode(data, Base64.DEFAULT)
                        val bitmap = BitmapFactory.decodeByteArray(bytes, 0, bytes.size)
                        if (bitmap != null) {
                            _uiState.value = _uiState.value.copy(
                                frameBitmap = bitmap,
                                frameWidth = bitmap.width,
                                frameHeight = bitmap.height,
                                frameCount = _uiState.value.frameCount + 1
                            )
                        }
                    }
                    val delayMs = 1000L / _uiState.value.fps
                    kotlinx.coroutines.delay(delayMs)
                } catch (e: Exception) {
                    _uiState.value = _uiState.value.copy(
                        error = "Stream error: ${e.message}",
                        isStreaming = false
                    )
                    break
                }
            }
        }
    }

    fun stopStream() {
        _uiState.value = _uiState.value.copy(isStreaming = false)
        streamJob?.cancel()
        streamJob = null
    }

    fun setQuality(quality: Int) {
        _uiState.value = _uiState.value.copy(quality = quality.coerceIn(10, 100))
    }

    fun setFps(fps: Int) {
        _uiState.value = _uiState.value.copy(fps = fps.coerceIn(1, 30))
    }

    fun sendInput(type: String, key: String = "", x: Int = 0, y: Int = 0, button: Int = 0, delta: Int = 0) {
        viewModelScope.launch {
            try {
                val params = org.json.JSONObject().apply {
                    put("type", type)
                    if (key.isNotEmpty()) put("key", key)
                    if (x != 0 || y != 0) {
                        put("x", x)
                        put("y", y)
                    }
                    if (button != 0) put("button", button)
                    if (delta != 0) put("delta", delta)
                }
                bppClient.sendRequest("screen.input", params)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.message ?: "Input failed")
            }
        }
    }

    fun reconnect() {
        streamJob?.cancel()
        _uiState.value = ScreenViewerUiState()
        connectAndCapture()
    }

    override fun onCleared() {
        streamJob?.cancel()
        viewModelScope.launch {
            bppClient.disconnect()
        }
    }
}
