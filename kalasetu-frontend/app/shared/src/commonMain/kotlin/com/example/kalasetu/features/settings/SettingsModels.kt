package com.example.kalasetu.features.settings

enum class ThemeMode(val label: String) {
    SYSTEM("System"),
    LIGHT("Light"),
    DARK("Dark")
}

sealed interface SettingsSection {
    data object Saved : SettingsSection
    data object Activity : SettingsSection
    data object ChangePassword : SettingsSection
    data object About : SettingsSection
}