package com.example.kalasetu.features.settings

import androidx.compose.runtime.Composable

interface SettingsStorage {
    fun getString(key: String, fallback: String): String
    fun putString(key: String, value: String)
}

@Composable
expect fun rememberSettingsStorage(): SettingsStorage

object SettingsStore {
    private const val KEY_THEME_MODE = "theme_mode"

    var storage: SettingsStorage? = null

    fun configure(storage: SettingsStorage) {
        this.storage = storage
    }

    fun getThemeMode(): ThemeMode {
        val stored = storage?.getString(KEY_THEME_MODE, ThemeMode.SYSTEM.name)
        return ThemeMode.entries.firstOrNull { it.name == stored } ?: ThemeMode.SYSTEM
    }

    fun setThemeMode(mode: ThemeMode) {
        storage?.putString(KEY_THEME_MODE, mode.name)
    }
}