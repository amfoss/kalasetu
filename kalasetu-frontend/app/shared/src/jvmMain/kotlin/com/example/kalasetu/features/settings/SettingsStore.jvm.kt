package com.example.kalasetu.features.settings

import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import java.util.prefs.Preferences

@Composable
actual fun rememberSettingsStorage(): SettingsStorage {
    return remember {
        val prefs = Preferences.userNodeForPackage(SettingsStorage::class.java)
        object : SettingsStorage {
            override fun getString(key: String, fallback: String): String =
                prefs.get(key, fallback)

            override fun putString(key: String, value: String) {
                prefs.put(key, value)
            }
        }
    }
}