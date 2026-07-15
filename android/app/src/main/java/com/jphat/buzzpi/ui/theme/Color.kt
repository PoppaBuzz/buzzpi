package com.jphat.buzzpi.ui.theme

import androidx.compose.ui.graphics.Color

// BuzzPi brand colors (from logo)
val BuzzPiBrandGreen = Color(0xFF669D2E)
val BuzzPiBrandGreenLight = Color(0xFF8BC34A)
val BuzzPiBrandGold = Color(0xFFE7BD14)
val BuzzPiBrandBlue = Color(0xFF015DE9)

// BuzzPi dark theme palette
val BuzzPiBg = Color(0xFF0A0E17)
val BuzzPiSurface = Color(0xFF111827)
val BuzzPiCard = Color(0xFF1A1F2E)
val BuzzPiCardHover = Color(0xFF1F2538)
val BuzzPiBorder = Color(0xFF1F2937)
val BuzzPiPrimary = BuzzPiBrandGreen
val BuzzPiPrimaryLight = BuzzPiBrandGreenLight
val BuzzPiAmber = Color(0xFFF59E0B)
val BuzzPiGold = BuzzPiBrandGold
val BuzzPiGreen = Color(0xFF22C55E)
val BuzzPiBlue = BuzzPiBrandBlue
val BuzzPiRed = Color(0xFFEF4444)
val BuzzPiText = Color(0xFFF1F5F9)
val BuzzPiTextSecondary = Color(0xFF94A3B8)
val BuzzPiTextTertiary = Color(0xFF64748B)
val BuzzPiSurfaceContainer = Color(0xFF1E293B)

// Light theme
val BuzzPiBgLight = Color(0xFFF8FAFC)
val BuzzPiSurfaceLight = Color(0xFFFFFFFF)
val BuzzPiCardLight = Color(0xFFF1F5F9)
val BuzzPiBorderLight = Color(0xFFE2E8F0)
val BuzzPiTextLight = Color(0xFF0F172A)
val BuzzPiTextSecondaryLight = Color(0xFF475569)

// Material 3 color scheme mapping
val BuzzPiDarkColorScheme = androidx.compose.material3.darkColorScheme(
    primary = BuzzPiPrimary,
    onPrimary = Color.White,
    primaryContainer = BuzzPiPrimary.copy(alpha = 0.15f),
    onPrimaryContainer = BuzzPiPrimaryLight,
    secondary = BuzzPiAmber,
    onSecondary = Color(0xFF1C1600),
    tertiary = BuzzPiGold,
    onTertiary = Color(0xFF1C1600),
    background = BuzzPiBg,
    onBackground = BuzzPiText,
    surface = BuzzPiSurface,
    onSurface = BuzzPiText,
    surfaceVariant = BuzzPiSurfaceContainer,
    onSurfaceVariant = BuzzPiTextSecondary,
    outline = BuzzPiBorder,
    outlineVariant = BuzzPiBorder,
    error = BuzzPiRed,
    onError = Color.White
)

val BuzzPiLightColorScheme = androidx.compose.material3.lightColorScheme(
    primary = BuzzPiPrimary,
    onPrimary = Color.White,
    primaryContainer = BuzzPiPrimary.copy(alpha = 0.1f),
    onPrimaryContainer = Color(0xFF33691E),
    secondary = BuzzPiAmber,
    onSecondary = Color.White,
    tertiary = BuzzPiGold,
    onTertiary = Color(0xFF1C1600),
    background = BuzzPiBgLight,
    onBackground = BuzzPiTextLight,
    surface = BuzzPiSurfaceLight,
    onSurface = BuzzPiTextLight,
    surfaceVariant = BuzzPiCardLight,
    onSurfaceVariant = BuzzPiTextSecondaryLight,
    outline = BuzzPiBorderLight,
    error = BuzzPiRed,
    onError = Color.White
)
