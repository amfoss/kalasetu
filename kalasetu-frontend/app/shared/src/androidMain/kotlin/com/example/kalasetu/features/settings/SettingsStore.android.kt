package com.example.kalasetu.features.settings

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.platform.LocalContext

@Composable
actual fun rememberSettingsStorage(): SettingsStorage {
    val context = LocalContext.current.applicationContext
    return remember(context) {
        val prefs = context.getSharedPreferences("kalasetu_settings", Context.MODE_PRIVATE)
        object : SettingsStorage {
            override fun getString(key: String, fallback: String): String =
                prefs.getString(key, fallback) ?: fallback

            override fun putString(key: String, value: String) {
                prefs.edit().putString(key, value).apply()
            }
        }
    }
}