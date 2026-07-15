package com.jphat.buzzpi.ui.screens.discovery.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.jphat.buzzpi.domain.model.Device

@Composable
fun PairingDialog(
    device: Device,
    onDismiss: () -> Unit,
    onPair: (Device, String) -> Unit,
    modifier: Modifier = Modifier
) {
    var pin by remember { mutableStateOf("") }
    var isPairing by remember { mutableStateOf(false) }

    AlertDialog(
        onDismissRequest = {
            if (!isPairing) onDismiss()
        },
        title = {
            Text(
                text = "Pair with ${device.friendlyName}",
                style = MaterialTheme.typography.headlineSmall
            )
        },
        text = {
            Column(modifier = modifier.fillMaxWidth()) {
                Text(
                    text = "Enter the PIN shown on your device's screen or terminal.",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.height(16.dp))
                OutlinedTextField(
                    value = pin,
                    onValueChange = { if (it.length <= 8) pin = it },
                    label = { Text("PIN") },
                    placeholder = { Text("XXXXXX") },
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(
                        keyboardType = KeyboardType.Number,
                        imeAction = ImeAction.Done
                    ),
                    keyboardActions = KeyboardActions(
                        onDone = {
                            if (pin.length >= 4) {
                                onPair(device, pin)
                            }
                        }
                    ),
                    enabled = !isPairing,
                    modifier = Modifier.fillMaxWidth()
                )
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "The PIN is a 6-digit code. It expires after 5 minutes.",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f),
                    textAlign = TextAlign.Start
                )
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    isPairing = true
                    onPair(device, pin)
                },
                enabled = pin.length >= 4 && !isPairing
            ) {
                Text(if (isPairing) "Pairing..." else "Pair")
            }
        },
        dismissButton = {
            TextButton(
                onClick = onDismiss,
                enabled = !isPairing
            ) {
                Text("Cancel")
            }
        }
    )
}
