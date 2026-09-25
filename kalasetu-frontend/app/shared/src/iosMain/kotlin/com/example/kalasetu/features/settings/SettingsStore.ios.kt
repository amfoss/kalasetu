package com.example.kalasetu.features.settings

import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import platform.Foundation.NSUserDefaults

@Composable
actual fun rememberSettingsStorage(): SettingsStorage {
    return remember {
        val defaults = NSUserDefaults.standardUserDefaults
        object : SettingsStorage {
            override fun getString(key: String, fallback: String): String =
                defaults.stringForKey(key) ?: fallback

            override fun putString(key: String, value: String) {
                defaults.setObject(value, forKey = key)
            }
        }
    }
}