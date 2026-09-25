package com.example.kalasetu.features.event

import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

val PurplePrimary = Color(0xFF7466F1)
val PurpleDark = Color(0xFF5548D9)
val LightPurpleBg @Composable get() = MaterialTheme.colorScheme.primaryContainer
val TextDark @Composable get() = MaterialTheme.colorScheme.onSurface
val TextGray @Composable get() = MaterialTheme.colorScheme.onSurfaceVariant
val BorderGray @Composable get() = MaterialTheme.colorScheme.outlineVariant
val FadedGray @Composable get() = MaterialTheme.colorScheme.onSurfaceVariant
val ErrorRed = Color(0xFFD32F2F)