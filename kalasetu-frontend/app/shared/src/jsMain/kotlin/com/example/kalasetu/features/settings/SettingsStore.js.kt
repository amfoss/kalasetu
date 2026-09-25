package com.example.kalasetu.features.settings

import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import kotlinx.browser.localStorage

@Composable
actual fun rememberSettingsStorage(): SettingsStorage {
    return remember {
        object : SettingsStorage {
            override fun getString(key: String, fallback: String): String =
                localStorage.getItem(key) ?: fallback

            override fun putString(key: String, value: String) {
                localStorage.setItem(key, value)
            }
        }
    }
}