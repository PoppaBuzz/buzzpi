package com.jphat.buzzpi.ui.screens.discovery

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Button
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
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import com.jphat.buzzpi.R
import com.jphat.buzzpi.domain.model.Device
import com.jphat.buzzpi.ui.components.BuzzPiTopBar
import com.jphat.buzzpi.ui.screens.discovery.components.DeviceCard
import com.jphat.buzzpi.ui.screens.discovery.components.PairingDialog

@Composable
fun DiscoveryScreen(
    onNavigateToPairing: (String) -> Unit,
    onNavigateToDashboard: (String) -> Unit,
    onNavigateToTerminal: (String) -> Unit,
    onNavigateToSettings: () -> Unit,
    viewModel: DiscoveryViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val selectedDevice by viewModel.selectedDevice.collectAsState()
    var showPairingDialog by remember { mutableStateOf(false) }
    var pairingDevice by remember { mutableStateOf<Device?>(null) }
    val snackbarHostState = remember { SnackbarHostState() }

    // Show errors
    LaunchedEffect(uiState.error) {
        uiState.error?.let { error ->
            snackbarHostState.showSnackbar(error)
        }
    }

    Scaffold(
        topBar = {
            BuzzPiTopBar(
                title = "BuzzPi",
                onSettings = onNavigateToSettings,
                actions = {
                    IconButton(onClick = { viewModel.refreshDiscovery() }) {
                        Icon(
                            imageVector = Icons.Filled.Refresh,
                            contentDescription = "Refresh"
                        )
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(hostState = snackbarHostState) }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            when {
                uiState.isScanning && uiState.devices.isEmpty() -> {
                    // Scanning state with logo
                    Column(
                        modifier = Modifier.align(Alignment.Center),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.ic_buzzpi_logo),
                            contentDescription = "BuzzPi",
                            modifier = Modifier.size(80.dp),
                            tint = MaterialTheme.colorScheme.primary
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        CircularProgressIndicator(
                            modifier = Modifier.size(24.dp),
                            color = MaterialTheme.colorScheme.primary,
                            strokeWidth = 2.dp
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            text = "Scanning for devices...",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
                uiState.devices.isEmpty() && !uiState.isScanning -> {
                    // Empty state
                    Column(
                        modifier = Modifier.align(Alignment.Center),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.ic_buzzpi_logo),
                            contentDescription = "BuzzPi",
                            modifier = Modifier.size(80.dp),
                            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f)
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            text = "No devices found",
                            style = MaterialTheme.typography.titleMedium,
                            color = MaterialTheme.colorScheme.onSurface
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Make sure your BuzzPi Runtime is running\non the same network.",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            textAlign = TextAlign.Center
                        )
                        Spacer(modifier = Modifier.height(24.dp))
                        Button(onClick = { viewModel.refreshDiscovery() }) {
                            Icon(
                                imageVector = Icons.Filled.Refresh,
                                contentDescription = null,
                                modifier = Modifier.size(18.dp)
                            )
                            Spacer(modifier = Modifier.size(8.dp))
                            Text("Scan Again")
                        }
                    }
                }
                else -> {
                    // Device list
                    LazyColumn(
                        contentPadding = PaddingValues(16.dp),
                        verticalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        // Paired devices section
                        if (uiState.pairedDevices.isNotEmpty()) {
                            item {
                                Text(
                                    text = "Paired Devices",
                                    style = MaterialTheme.typography.titleSmall,
                                    color = MaterialTheme.colorScheme.primary,
                                    modifier = Modifier.padding(top = 8.dp, bottom = 4.dp)
                                )
                            }
                            items(uiState.pairedDevices, key = { it.deviceId }) { device ->
                                DeviceCard(
                                    device = device,
                                    isPaired = true,
                                    onClick = { onNavigateToDashboard(device.deviceId) }
                                )
                            }
                        }

                        // Discovered devices section
                        if (uiState.devices.isNotEmpty()) {
                            item {
                                Text(
                                    text = "Discovered Devices",
                                    style = MaterialTheme.typography.titleSmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    modifier = Modifier.padding(
                                        top = if (uiState.pairedDevices.isNotEmpty()) 16.dp else 8.dp,
                                        bottom = 4.dp
                                    )
                                )
                            }
                            items(uiState.devices, key = { it.deviceId }) { device ->
                                DeviceCard(
                                    device = device,
                                    onClick = {
                                        pairingDevice = device
                                        showPairingDialog = true
                                    }
                                )
                            }
                        }
                    }
                }
            }

            // Scanning indicator at top
            if (uiState.isScanning && uiState.devices.isNotEmpty()) {
                CircularProgressIndicator(
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .padding(top = 8.dp)
                        .size(16.dp),
                    strokeWidth = 2.dp
                )
            }
        }
    }

    // Pairing dialog
    if (showPairingDialog && pairingDevice != null) {
        PairingDialog(
            device = pairingDevice!!,
            onDismiss = {
                showPairingDialog = false
                pairingDevice = null
            },
            onPair = { device, pin ->
                viewModel.pairWithDevice(device)
                showPairingDialog = false
            }
        )
    }
}
