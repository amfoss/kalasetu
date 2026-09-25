package com.example.kalasetu.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

val SelectedPurple = Color(0xFF836AE0)

val SubtitleGray: Color
    @Composable get() = MaterialTheme.colorScheme.onSurfaceVariant
val UnselectedBorder: Color
    @Composable get() = MaterialTheme.colorScheme.outlineVariant

private val LightColors = lightColorScheme(
    primary = Color(0xFF7466F1),
    onPrimary = Color.White,
    primaryContainer = Color(0xFFEDE7F6),
    onPrimaryContainer = Color(0xFF221346),
    secondary = Color(0xFFD4A843),
    surface = Color.White,
    background = Color(0xFFFDFBFF),
    onBackground = Color(0xFF1C1B1F),
    onSurface = Color(0xFF1C1B1F),
    onSurfaceVariant = Color(0xFF49454F),
    outline = Color(0xFF79747E),
    outlineVariant = Color(0xFFE0E0E0),
)

private val DarkColors = darkColorScheme(
    primary = Color(0xFFB8B0FF),
    onPrimary = Color(0xFF241B6B),
    primaryContainer = Color(0xFF3E349E),
    onPrimaryContainer = Color(0xFFE5E0FF),
    secondary = Color(0xFFE3C05C),
    surface = Color(0xFF16141A),
    background = Color(0xFF121016),
    onBackground = Color(0xFFE6E1E5),
    onSurface = Color(0xFFE6E1E5),
    onSurfaceVariant = Color(0xFFCAC4CF),
    outline = Color(0xFF938F99),
    outlineVariant = Color(0xFF49454F),
)

@Composable
fun KalasetuTheme(
    darkTheme: Boolean = false,
    content: @Composable () -> Unit
) {
    MaterialTheme(
        colorScheme = if (darkTheme) DarkColors else LightColors,
        content = content
    )
}