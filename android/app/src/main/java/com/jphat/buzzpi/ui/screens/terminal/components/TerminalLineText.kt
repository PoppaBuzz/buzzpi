package com.jphat.buzzpi.ui.screens.terminal.components

import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.BasicText
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import com.jphat.buzzpi.domain.model.StyledSegment
import com.jphat.buzzpi.domain.model.TerminalLine

@Composable
fun TerminalLineText(
    line: TerminalLine,
    modifier: Modifier = Modifier
) {
    val annotatedString = buildAnnotatedString {
        if (line.segments.isNotEmpty()) {
            for (segment in line.segments) {
                withStyle(style = segment.toSpanStyle()) {
                    append(segment.text)
                }
            }
        } else if (line.text.isNotEmpty()) {
            withStyle(
                style = SpanStyle(
                    fontFamily = FontFamily.Monospace,
                    fontSize = MaterialTheme.typography.bodyMedium.fontSize,
                    color = MaterialTheme.colorScheme.onSurface
                )
            ) {
                append(line.text)
            }
        }
    }

    BasicText(
        text = annotatedString,
        modifier = modifier.padding(horizontal = 8.dp, vertical = 1.dp)
    )
}

@Composable
private fun StyledSegment.toSpanStyle(): SpanStyle {
    val fgColor = if (fgColor != 0) Color(fgColor) else MaterialTheme.colorScheme.onSurface
    return SpanStyle(
        fontFamily = FontFamily.Monospace,
        fontSize = MaterialTheme.typography.bodyMedium.fontSize,
        color = fgColor,
        fontWeight = if (bold) FontWeight.Bold else FontWeight.Normal,
        fontStyle = if (italic) FontStyle.Italic else FontStyle.Normal,
        textDecoration = if (underline) TextDecoration.Underline else null
    )
}
