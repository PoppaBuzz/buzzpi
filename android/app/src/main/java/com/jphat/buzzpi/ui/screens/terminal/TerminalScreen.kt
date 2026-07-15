package com.jphat.buzzpi.ui.screens.terminal

import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ClearAll
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import com.jphat.buzzpi.data.bpp.ConnectionState
import com.jphat.buzzpi.ui.components.BuzzPiTopBar
import com.jphat.buzzpi.ui.components.ConnectionIndicator
import com.jphat.buzzpi.ui.screens.terminal.components.TerminalInputBar
import com.jphat.buzzpi.ui.screens.terminal.components.TerminalLineText

@Composable
fun TerminalScreen(
    onNavigateBack: () -> Unit,
    viewModel: TerminalViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val scrollState = rememberScrollState()
    val horizontalScrollState = rememberScrollState()
    val snackbarHostState = remember { SnackbarHostState() }

    // Auto-scroll when new output arrives
    LaunchedEffect(uiState.terminalState.lines.size) {
        scrollState.animateScrollTo(scrollState.maxValue)
    }

    // Show errors
    LaunchedEffect(uiState.error) {
        uiState.error?.let { error ->
            snackbarHostState.showSnackbar(error)
        }
    }

    Scaffold(
        topBar = {
            BuzzPiTopBar(
                title = "Terminal",
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
                    IconButton(onClick = { viewModel.clearTerminal() }) {
                        Icon(Icons.Filled.ClearAll, "Clear")
                    }
                    IconButton(onClick = { viewModel.reconnect() }) {
                        Icon(Icons.Filled.Refresh, "Reconnect")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(hostState = snackbarHostState) }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Terminal output area
            Box(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .background(Color(0xFF0D1117))
            ) {
                if (uiState.isConnecting) {
                    Column(
                        modifier = Modifier.align(Alignment.Center),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(32.dp),
                            color = Color(0xFF22C55E)
                        )
                        Text(
                            text = "Connecting to terminal...",
                            color = Color(0xFF8B949E),
                            modifier = Modifier.padding(top = 12.dp)
                        )
                    }
                } else if (uiState.terminalState.lines.isEmpty()) {
                    Text(
                        text = "Terminal ready. Type a command to begin.",
                        color = Color(0xFF8B949E),
                        modifier = Modifier
                            .align(Alignment.Center)
                            .padding(16.dp)
                    )
                } else {
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .verticalScroll(scrollState)
                            .horizontalScroll(horizontalScrollState)
                            .padding(vertical = 8.dp)
                    ) {
                        uiState.terminalState.lines.forEach { line ->
                            TerminalLineText(line = line)
                        }
                    }
                }
            }

            // Input bar
            TerminalInputBar(
                input = uiState.inputBuffer,
                onInputChange = { viewModel.updateInputBuffer(it) },
                onSend = { viewModel.sendInput(it) },
                isConnected = uiState.isConnected
            )
        }
    }
}
