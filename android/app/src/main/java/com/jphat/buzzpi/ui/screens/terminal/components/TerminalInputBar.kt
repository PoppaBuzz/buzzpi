package com.jphat.buzzpi.ui.screens.terminal.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.defaultMinSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp

@Composable
fun TerminalInputBar(
    input: String,
    onInputChange: (String) -> Unit,
    onSend: (String) -> Unit,
    onSendKey: (ByteArray) -> Unit,
    isConnected: Boolean,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier.fillMaxWidth()) {
        // Special keys toolbar
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState())
                .padding(horizontal = 8.dp, vertical = 4.dp),
            horizontalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            SpecialKeyButton("Esc", byteArrayOf(0x1B), onSendKey, isConnected)
            SpecialKeyButton("Tab", byteArrayOf(0x09), onSendKey, isConnected)
            SpecialKeyButton("Ctrl+C", byteArrayOf(0x03), onSendKey, isConnected)
            SpecialKeyButton("Ctrl+D", byteArrayOf(0x04), onSendKey, isConnected)
            SpecialKeyButton("Ctrl+Z", byteArrayOf(0x1A), onSendKey, isConnected)
            SpecialKeyButton("↑", byteArrayOf(0x1B, 0x5B, 0x41), onSendKey, isConnected)
            SpecialKeyButton("↓", byteArrayOf(0x1B, 0x5B, 0x42), onSendKey, isConnected)
            SpecialKeyButton("→", byteArrayOf(0x1B, 0x5B, 0x43), onSendKey, isConnected)
            SpecialKeyButton("←", byteArrayOf(0x1B, 0x5B, 0x44), onSendKey, isConnected)
            SpecialKeyButton("Home", byteArrayOf(0x1B, 0x5B, 0x48), onSendKey, isConnected)
            SpecialKeyButton("End", byteArrayOf(0x1B, 0x5B, 0x46), onSendKey, isConnected)
            SpecialKeyButton("PgUp", byteArrayOf(0x1B, 0x5B, 0x35, 0x7E), onSendKey, isConnected)
            SpecialKeyButton("PgDn", byteArrayOf(0x1B, 0x5B, 0x36, 0x7E), onSendKey, isConnected)
        }

        // Input field
        OutlinedTextField(
            value = input,
            onValueChange = onInputChange,
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 8.dp, vertical = 8.dp),
            placeholder = {
                Text(
                    text = if (isConnected) "$ " else "Disconnected...",
                    fontFamily = FontFamily.Monospace,
                    style = MaterialTheme.typography.bodyMedium
                )
            },
            textStyle = MaterialTheme.typography.bodyMedium.copy(
                fontFamily = FontFamily.Monospace
            ),
            singleLine = true,
            enabled = isConnected,
            shape = RoundedCornerShape(12.dp),
            colors = OutlinedTextFieldDefaults.colors(
                focusedBorderColor = MaterialTheme.colorScheme.primary,
                unfocusedBorderColor = MaterialTheme.colorScheme.outline.copy(alpha = 0.3f),
                cursorColor = MaterialTheme.colorScheme.primary
            ),
            keyboardOptions = KeyboardOptions(
                imeAction = ImeAction.Send
            ),
            keyboardActions = KeyboardActions(
                onSend = {
                    if (input.isNotBlank()) {
                        onSend(input.trimEnd())
                    }
                }
            ),
            trailingIcon = {
                IconButton(
                    onClick = { onSend(input.trimEnd()) },
                    enabled = input.isNotBlank() && isConnected
                ) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.Send,
                        contentDescription = "Send",
                        tint = if (input.isNotBlank() && isConnected)
                            MaterialTheme.colorScheme.primary
                        else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f)
                    )
                }
            }
        )
    }
}

@Composable
private fun SpecialKeyButton(
    label: String,
    keyBytes: ByteArray,
    onSendKey: (ByteArray) -> Unit,
    enabled: Boolean
) {
    Box(
        modifier = Modifier
            .height(32.dp)
            .defaultMinSize(minWidth = 40.dp)
            .clip(RoundedCornerShape(6.dp))
            .background(
                if (enabled) MaterialTheme.colorScheme.surfaceVariant
                else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
            )
            .clickable(enabled = enabled) { onSendKey(keyBytes) }
            .padding(horizontal = 8.dp, vertical = 4.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            fontFamily = FontFamily.Monospace,
            color = if (enabled) MaterialTheme.colorScheme.onSurfaceVariant
            else MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.3f)
        )
    }
}
