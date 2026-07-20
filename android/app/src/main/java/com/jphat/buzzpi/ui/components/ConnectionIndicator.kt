package com.jphat.buzzpi.ui.components

import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.jphat.buzzpi.data.bpp.ConnectionState

@Composable
fun ConnectionIndicator(
    state: ConnectionState,
    modifier: Modifier = Modifier,
    size: Dp = 10.dp
) {
    val color by animateColorAsState(
        targetValue = when (state) {
            ConnectionState.CONNECTED -> Color(0xFF22C55E)
            ConnectionState.CONNECTING, ConnectionState.RECONNECTING -> Color(0xFFF59E0B)
            ConnectionState.DISCONNECTED, ConnectionState.ERROR -> Color(0xFFEF4444)
        },
        label = "connectionColor"
    )

    Box(
        modifier = modifier
            .size(size)
            .clip(CircleShape)
            .background(color)
    )
}
