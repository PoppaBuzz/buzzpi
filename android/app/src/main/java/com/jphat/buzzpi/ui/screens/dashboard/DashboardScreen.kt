package com.jphat.buzzpi.ui.screens.dashboard

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Memory
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Storage
import androidx.compose.material.icons.filled.Terminal
import androidx.compose.material.icons.filled.DesktopWindows
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ElevatedButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import com.jphat.buzzpi.data.bpp.ConnectionState
import com.jphat.buzzpi.data.bpp.CPUStats
import com.jphat.buzzpi.data.bpp.DiskStats
import com.jphat.buzzpi.data.bpp.MemoryStats
import com.jphat.buzzpi.ui.components.BuzzPiTopBar
import com.jphat.buzzpi.ui.components.ConnectionIndicator

@Composable
fun DashboardScreen(
    onNavigateToTerminal: (String) -> Unit,
    onNavigateToScreenViewer: (String) -> Unit,
    onNavigateToFileManager: (String) -> Unit,
    onNavigateBack: () -> Unit,
    viewModel: DashboardViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            BuzzPiTopBar(
                title = uiState.device?.friendlyName ?: "Dashboard",
                showBack = true,
                onBack = {
                    viewModel.disconnect()
                    onNavigateBack()
                },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Filled.Refresh, "Refresh")
                    }
                }
            )
        }
    ) { padding ->
        when {
            uiState.isLoading -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    CircularProgressIndicator(modifier = Modifier.size(48.dp))
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        "Loading device info...",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            uiState.device == null -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Text("Device not found", style = MaterialTheme.typography.titleMedium)
                    Text(
                        uiState.error ?: "Unknown error",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            else -> {
                val device = uiState.device!!
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .verticalScroll(rememberScrollState())
                        .padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    // Status card
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp),
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.surfaceVariant
                        )
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                ConnectionIndicator(
                                    state = if (device.isOnline) ConnectionState.CONNECTED
                                    else ConnectionState.DISCONNECTED,
                                    size = 14.dp
                                )
                                Spacer(modifier = Modifier.width(10.dp))
                                Text(
                                    text = if (device.isOnline) "Online" else "Offline",
                                    style = MaterialTheme.typography.titleMedium
                                )
                            }
                            Spacer(modifier = Modifier.height(8.dp))
                            device.platform.let { platform ->
                                if (platform.isNotEmpty()) {
                                    Text(
                                        "Platform: $platform",
                                        style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant
                                    )
                                }
                            }
                            if (device.runtimeVersion.isNotEmpty()) {
                                Text(
                                    "Runtime: v${device.runtimeVersion}",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            }
                        }
                    }

                    // Quick actions
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp),
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.surfaceVariant
                        )
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(
                                "Quick Actions",
                                style = MaterialTheme.typography.titleSmall,
                                color = MaterialTheme.colorScheme.primary
                            )
                            Spacer(modifier = Modifier.height(12.dp))
                            ElevatedButton(
                                onClick = { onNavigateToTerminal(device.deviceId) },
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Icon(Icons.Filled.Terminal, contentDescription = null, modifier = Modifier.size(18.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("Open Terminal")
                            }
                            Spacer(modifier = Modifier.height(8.dp))
                            ElevatedButton(
                                onClick = { onNavigateToScreenViewer(device.deviceId) },
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Icon(Icons.Filled.DesktopWindows, contentDescription = null, modifier = Modifier.size(18.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("Screen Viewer")
                            }
                            Spacer(modifier = Modifier.height(8.dp))
                            ElevatedButton(
                                onClick = { onNavigateToFileManager(device.deviceId) },
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Icon(Icons.Filled.Folder, contentDescription = null, modifier = Modifier.size(18.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("File Manager")
                            }
                        }
                    }

                    // Device info
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp),
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.surfaceVariant
                        )
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(
                                "Device Info",
                                style = MaterialTheme.typography.titleSmall,
                                color = MaterialTheme.colorScheme.primary
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            InfoRow("Device ID", device.deviceId)
                            device.friendlyName.let { if (it.isNotEmpty()) InfoRow("Name", it) }
                            device.platform.let { if (it.isNotEmpty()) InfoRow("Platform", it) }
                            device.runtimeVersion.let { if (it.isNotEmpty()) InfoRow("Runtime", it) }
                        }
                    }

                    // Live stats
                    uiState.cpu?.let { cpu -> CpuStatsCard(cpu) }
                    uiState.memory?.let { mem -> MemoryStatsCard(mem) }
                    uiState.storage?.let { disks -> StorageStatsCard(disks) }
                }
            }
        }
    }
}

@Composable
private fun StatCard(
    icon: ImageVector,
    title: String,
    content: @Composable () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    icon, contentDescription = null,
                    modifier = Modifier.size(18.dp),
                    tint = MaterialTheme.colorScheme.primary
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    title,
                    style = MaterialTheme.typography.titleSmall,
                    color = MaterialTheme.colorScheme.primary
                )
            }
            Spacer(modifier = Modifier.height(10.dp))
            content()
        }
    }
}

@Composable
private fun CpuStatsCard(cpu: CPUStats) {
    StatCard(icon = Icons.Filled.Info, title = "CPU") {
        StatBar("Usage", cpu.usagePercent, "%.1f%%".format(cpu.usagePercent))
        Text(
            "Temp: %.1f\u00B0C".format(cpu.temperatureCelsius),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            "Freq: %d MHz".format(cpu.frequencyMhz),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
private fun MemoryStatsCard(mem: MemoryStats) {
    StatCard(icon = Icons.Filled.Memory, title = "Memory") {
        StatBar("Used", mem.percent, "%.1f%%".format(mem.percent))
        InfoRow("Total", "%,d MB".format(mem.totalMb))
        InfoRow("Used", "%,d MB".format(mem.usedMb))
        InfoRow("Available", "%,d MB".format(mem.availableMb))
    }
}

@Composable
private fun StorageStatsCard(disks: List<DiskStats>) {
    StatCard(icon = Icons.Filled.Storage, title = "Storage") {
        disks.forEach { disk ->
            Text(
                disk.mount,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurface
            )
            StatBar("", disk.percent, "%.1f%% \u2014 %,d / %,d MB".format(disk.percent, disk.usedMb, disk.totalMb))
            Spacer(modifier = Modifier.height(6.dp))
        }
    }
}

@Composable
private fun StatBar(label: String, value: Double, display: String) {
    Column(modifier = Modifier.padding(vertical = 4.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            if (label.isNotEmpty()) {
                Text(
                    label, style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            Text(
                display, style = MaterialTheme.typography.bodySmall,
                fontFamily = FontFamily.Monospace,
                color = MaterialTheme.colorScheme.onSurface
            )
        }
        Spacer(modifier = Modifier.height(4.dp))

        val progress = (value / 100.0).coerceIn(0.0, 1.0)
        val barColor = when {
            value >= 90 -> Color(0xFFE53935)
            value >= 70 -> Color(0xFFFB8C00)
            else -> Color(0xFF43A047)
        }
        LinearProgressIndicator(
            progress = { progress.toFloat() },
            modifier = Modifier
                .fillMaxWidth()
                .height(6.dp),
            color = barColor,
            trackColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
        )
    }
}

@Composable
private fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 2.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            label, style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            value, style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            color = MaterialTheme.colorScheme.onSurface
        )
    }
}
