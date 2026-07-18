package com.jphat.buzzpi.ui.screens.screenviewer

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.gestures.detectTransformGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CropFree
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Screenshot
import androidx.compose.material.icons.filled.Stop
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ElevatedButton
import androidx.compose.material3.ElevatedFilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Slider
import androidx.compose.material3.SliderDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import com.jphat.buzzpi.data.bpp.ConnectionState
import com.jphat.buzzpi.ui.components.BuzzPiTopBar
import com.jphat.buzzpi.ui.components.ConnectionIndicator

@Composable
fun ScreenViewerScreen(
    onNavigateBack: () -> Unit,
    viewModel: ScreenViewerViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    var scale by remember { mutableFloatStateOf(1f) }
    var offsetX by remember { mutableFloatStateOf(0f) }
    var offsetY by remember { mutableFloatStateOf(0f) }
    var showControls by remember { mutableStateOf(true) }

    Scaffold(
        topBar = {
            if (showControls) {
                BuzzPiTopBar(
                    title = "Remote Desktop",
                    showBack = true,
                    onBack = onNavigateBack,
                    actions = {
                        ConnectionIndicator(
                            state = when {
                                uiState.isConnected -> ConnectionState.CONNECTED
                                uiState.isConnecting -> ConnectionState.CONNECTING
                                else -> ConnectionState.DISCONNECTED
                            },
                            size = 14.dp
                        )
                        IconButton(onClick = { viewModel.captureScreenshot() }) {
                            Icon(Icons.Filled.Screenshot, "Screenshot")
                        }
                        IconButton(onClick = { viewModel.reconnect() }) {
                            Icon(Icons.Filled.Refresh, "Reconnect")
                        }
                    }
                )
            }
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .background(Color(0xFF0D1117))
                .pointerInput(Unit) {
                    detectTransformGestures { _, pan, zoom, _ ->
                        scale = (scale * zoom).coerceIn(0.5f, 5f)
                        offsetX += pan.x
                        offsetY += pan.y
                    }
                }
                .pointerInput(Unit) {
                    awaitEachGesture {
                        val down = awaitFirstDown()
                        val x = (down.position.x / scale - offsetX / scale).toInt()
                        val y = (down.position.y / scale - offsetY / scale).toInt()
                        val w = uiState.frameWidth.coerceAtLeast(1)
                        val h = uiState.frameHeight.coerceAtLeast(1)
                        val screenW = 1280
                        val screenH = 720
                        val mappedX = (x.toFloat() / w * screenW).toInt().coerceIn(0, screenW)
                        val mappedY = (y.toFloat() / h * screenH).toInt().coerceIn(0, screenH)
                        viewModel.sendInput("mouse_click", x = mappedX, y = mappedY, button = 1)
                    }
                }
        ) {
            when {
                uiState.isConnecting -> {
                    Column(
                        modifier = Modifier.align(Alignment.Center),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(48.dp),
                            color = Color(0xFF22C55E)
                        )
                        Text(
                            text = "Connecting to device...",
                            color = Color(0xFF8B949E),
                            modifier = Modifier.padding(top = 12.dp)
                        )
                    }
                }
                uiState.frameBitmap != null -> {
                    Image(
                        bitmap = uiState.frameBitmap!!.asImageBitmap(),
                        contentDescription = "Remote screen",
                        contentScale = ContentScale.Fit,
                        modifier = Modifier
                            .fillMaxSize()
                            .graphicsLayer(
                                scaleX = scale,
                                scaleY = scale,
                                translationX = offsetX,
                                translationY = offsetY
                            )
                    )

                    if (showControls) {
                        Column(
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth()
                                .background(Color(0xCC000000))
                                .padding(12.dp)
                        ) {
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Text(
                                    "Quality: ${uiState.quality}%",
                                    color = Color.White,
                                    style = MaterialTheme.typography.bodySmall
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                Slider(
                                    value = uiState.quality.toFloat(),
                                    onValueChange = { viewModel.setQuality(it.toInt()) },
                                    valueRange = 10f..100f,
                                    modifier = Modifier.weight(1f),
                                    colors = SliderDefaults.colors(
                                        thumbColor = Color(0xFF22C55E),
                                        activeTrackColor = Color(0xFF22C55E)
                                    )
                                )
                            }

                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                ElevatedFilterChip(
                                    selected = uiState.isStreaming,
                                    onClick = {
                                        if (uiState.isStreaming) viewModel.stopStream()
                                        else viewModel.startStream()
                                    },
                                    label = { Text(if (uiState.isStreaming) "Stop Stream" else "Start Stream") },
                                    leadingIcon = {
                                        Icon(
                                            if (uiState.isStreaming) Icons.Filled.Stop else Icons.Filled.PlayArrow,
                                            contentDescription = null,
                                            modifier = Modifier.size(18.dp)
                                        )
                                    },
                                    colors = FilterChipDefaults.elevatedFilterChipColors(
                                        selectedContainerColor = Color(0xFF22C55E).copy(alpha = 0.2f)
                                    )
                                )

                                Spacer(modifier = Modifier.weight(1f))

                                Text(
                                    "Frame: ${uiState.frameCount}",
                                    color = Color(0xFF8B949E),
                                    style = MaterialTheme.typography.bodySmall
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                Text(
                                    "${uiState.frameWidth}x${uiState.frameHeight}",
                                    color = Color(0xFF8B949E),
                                    style = MaterialTheme.typography.bodySmall
                                )
                            }
                        }
                    }
                }
                else -> {
                    Column(
                        modifier = Modifier.align(Alignment.Center),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Icon(
                            Icons.Filled.CropFree,
                            contentDescription = null,
                            modifier = Modifier.size(64.dp),
                            tint = Color(0xFF8B949E)
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            text = uiState.error ?: "Connecting to Pi display...",
                            color = Color(0xFF8B949E)
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        ElevatedButton(onClick = { viewModel.captureScreenshot() }) {
                            Text("Capture Screenshot")
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        ElevatedFilterChip(
                            selected = uiState.isStreaming,
                            onClick = {
                                if (uiState.isStreaming) viewModel.stopStream()
                                else viewModel.startStream()
                            },
                            label = { Text(if (uiState.isStreaming) "Stop Stream" else "Start Live Stream") },
                            leadingIcon = {
                                Icon(
                                    if (uiState.isStreaming) Icons.Filled.Stop else Icons.Filled.PlayArrow,
                                    contentDescription = null,
                                    modifier = Modifier.size(18.dp)
                                )
                            },
                            colors = FilterChipDefaults.elevatedFilterChipColors(
                                selectedContainerColor = Color(0xFF22C55E).copy(alpha = 0.2f)
                            )
                        )
                    }
                }
            }
        }
    }
}
