package com.jphat.buzzpi.ui.screens.pairing

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Error
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import com.jphat.buzzpi.R
import com.jphat.buzzpi.ui.components.BuzzPiTopBar

@Composable
fun PairingScreen(
    onNavigateBack: () -> Unit,
    onPairingComplete: (String) -> Unit,
    viewModel: PairingViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            BuzzPiTopBar(
                title = "Pair Device",
                showBack = true,
                onBack = onNavigateBack
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            when (uiState.stage) {
                PairingStage.INITIATING -> {
                    InitiatingContent()
                }
                PairingStage.WAITING_FOR_CONFIRMATION -> {
                    WaitingForConfirmationContent(uiState)
                }
                PairingStage.VERIFYING -> {
                    VerifyingContent()
                }
                PairingStage.COMPLETED -> {
                    CompletedContent(
                        deviceId = "" // Get from viewModel
                    ) {
                        // Navigate to terminal after short delay
                        onPairingComplete("")
                    }
                }
                PairingStage.FAILED -> {
                    FailedContent(
                        error = uiState.error,
                        onRetry = { viewModel.retry() },
                        onNavigateBack = onNavigateBack
                    )
                }
            }
        }
    }
}

@Composable
private fun InitiatingContent() {
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
        text = "Initiating pairing...",
        style = MaterialTheme.typography.titleMedium,
        color = MaterialTheme.colorScheme.onSurface
    )
    Spacer(modifier = Modifier.height(8.dp))
    Text(
        text = "Connecting to the device and requesting a pairing session.",
        style = MaterialTheme.typography.bodyMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = TextAlign.Center
    )
}

@Composable
private fun WaitingForConfirmationContent(uiState: PairingUiState) {
    Icon(
        painter = painterResource(R.drawable.ic_buzzpi_logo),
        contentDescription = "BuzzPi",
        modifier = Modifier.size(80.dp),
        tint = MaterialTheme.colorScheme.primary
    )
    Spacer(modifier = Modifier.height(16.dp))
    CircularProgressIndicator(
        modifier = Modifier.size(24.dp),
        color = MaterialTheme.colorScheme.tertiary,
        strokeWidth = 2.dp
    )
    Spacer(modifier = Modifier.height(16.dp))
    Text(
        text = "Waiting for confirmation",
        style = MaterialTheme.typography.titleMedium,
        color = MaterialTheme.colorScheme.onSurface
    )
    Spacer(modifier = Modifier.height(8.dp))
    Text(
        text = "Check the device screen. A pairing code should appear.\nConfirm the pairing on the device to continue.",
        style = MaterialTheme.typography.bodyMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(24.dp))

    // Show pairing session info if available
    val session = uiState.session
    if (session != null && session.pin.isNotEmpty()) {
        Text(
            text = "Device PIN: ${session.pin}",
            style = MaterialTheme.typography.headlineLarge,
            color = MaterialTheme.colorScheme.primary
        )
    }

    Spacer(modifier = Modifier.height(32.dp))
    Text(
        text = "This should match the code on your device's display.",
        style = MaterialTheme.typography.labelMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f),
        textAlign = TextAlign.Center
    )
}

@Composable
private fun VerifyingContent() {
    val rotation by animateFloatAsState(
        targetValue = 360f,
        label = "rotation"
    )
    CircularProgressIndicator(
        modifier = Modifier
            .size(48.dp)
            .rotate(rotation),
        color = MaterialTheme.colorScheme.primary
    )
    Spacer(modifier = Modifier.height(24.dp))
    Text(
        text = "Verifying pairing...",
        style = MaterialTheme.typography.titleMedium,
        color = MaterialTheme.colorScheme.onSurface
    )
}

@Composable
private fun CompletedContent(
    deviceId: String,
    onNavigateToTerminal: (String) -> Unit
) {
    Icon(
        imageVector = Icons.Filled.CheckCircle,
        contentDescription = null,
        modifier = Modifier.size(80.dp),
        tint = MaterialTheme.colorScheme.primary
    )
    Spacer(modifier = Modifier.height(24.dp))
    Text(
        text = "Device Paired!",
        style = MaterialTheme.typography.headlineMedium,
        color = MaterialTheme.colorScheme.onSurface
    )
    Spacer(modifier = Modifier.height(8.dp))
    Text(
        text = "Your BuzzPi Runtime has been paired successfully.\nYou can now connect to the terminal.",
        style = MaterialTheme.typography.bodyMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(32.dp))
    Button(
        onClick = { onNavigateToTerminal(deviceId) },
        modifier = Modifier.fillMaxWidth()
    ) {
        Text("Open Terminal")
    }
}

@Composable
private fun FailedContent(
    error: String?,
    onRetry: () -> Unit,
    onNavigateBack: () -> Unit
) {
    Icon(
        imageVector = Icons.Filled.Error,
        contentDescription = null,
        modifier = Modifier.size(64.dp),
        tint = MaterialTheme.colorScheme.error
    )
    Spacer(modifier = Modifier.height(24.dp))
    Text(
        text = "Pairing Failed",
        style = MaterialTheme.typography.titleLarge,
        color = MaterialTheme.colorScheme.onSurface
    )
    if (error != null) {
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = error,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
    }
    Spacer(modifier = Modifier.height(24.dp))
    Button(
        onClick = onRetry,
        modifier = Modifier.fillMaxWidth()
    ) {
        Text("Retry")
    }
    Spacer(modifier = Modifier.height(8.dp))
    TextButton(onClick = { onNavigateBack() }) {
        Text("Go Back")
    }
}
