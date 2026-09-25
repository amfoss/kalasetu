package com.example.kalasetu.features.settings

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.outlined.Logout
import androidx.compose.material.icons.filled.ArrowDropDown
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.outlined.BookmarkBorder
import androidx.compose.material.icons.outlined.DarkMode
import androidx.compose.material.icons.outlined.History
import androidx.compose.material.icons.outlined.Info
import androidx.compose.material.icons.outlined.Lock
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil3.compose.AsyncImage
import com.example.kalasetu.features.auth.AuthStore
import com.example.kalasetu.features.profile.toInitials
import com.example.kalasetu.navigation.BackHandler

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    userName: String?,
    userEmail: String?,
    userAvatarUrl: String?,
    userAvatarBytes: ByteArray?,
    themeMode: ThemeMode,
    onThemeModeChange: (ThemeMode) -> Unit,
    onNavigateToProfile: () -> Unit,
    onBack: () -> Unit,
    onMenuClick: () -> Unit,
    onLoggedOut: () -> Unit
) {
    var section by remember { mutableStateOf<SettingsSection?>(null) }
    var showLogoutDialog by remember { mutableStateOf(false) }

    BackHandler {
        if (section != null) section = null else onBack()
    }

    val active = section
    if (active != null) {
        when (active) {
            SettingsSection.Saved -> SavedScreen(onBack = { section = null })
            SettingsSection.Activity -> ActivityScreen(
                userId = AuthStore.userId?.toString() ?: "",
                onBack = { section = null }
            )
            SettingsSection.ChangePassword -> ChangePasswordScreen(onBack = { section = null })
            SettingsSection.About -> AboutScreen(onBack = { section = null })
        }
        return
    }

    Scaffold(
        containerColor = MaterialTheme.colorScheme.background,
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("Settings", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    SettingsAvatarButton(
                        avatarUrl = userAvatarUrl,
                        avatarBytes = userAvatarBytes,
                        userName = userName,
                        onClick = onNavigateToProfile
                    )
                },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState())
        ) {
            SettingsGroupLabel("Preferences")
            ThemeSettingRow(
                themeMode = themeMode,
                onThemeModeChange = onThemeModeChange
            )

            SettingsGroupLabel("Content")
            SettingsRow(
                icon = Icons.Outlined.BookmarkBorder,
                label = "Saved",
                onClick = { section = SettingsSection.Saved }
            )
            SettingsRow(
                icon = Icons.Outlined.History,
                label = "Your Activity",
                onClick = { section = SettingsSection.Activity }
            )

            SettingsGroupLabel("Account")
            SettingsRow(
                icon = Icons.Outlined.Lock,
                label = "Change Password",
                onClick = { section = SettingsSection.ChangePassword }
            )
            SettingsRow(
                icon = Icons.AutoMirrored.Outlined.Logout,
                label = "Log Out",
                tint = MaterialTheme.colorScheme.error,
                onClick = { showLogoutDialog = true }
            )

            SettingsGroupLabel("App")
            SettingsRow(
                icon = Icons.Outlined.Info,
                label = "About",
                onClick = { section = SettingsSection.About }
            )

            Spacer(modifier = Modifier.height(32.dp))
        }
    }

    if (showLogoutDialog) {
        AlertDialog(
            onDismissRequest = { showLogoutDialog = false },
            title = { Text("Log out") },
            text = { Text("Are you sure you want to log out?") },
            confirmButton = {
                TextButton(
                    onClick = {
                        showLogoutDialog = false
                        onLoggedOut()
                    }
                ) {
                    Text("Log Out", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { showLogoutDialog = false }) {
                    Text("Cancel")
                }
            }
        )
    }
}

@Composable
private fun ThemeSettingRow(
    themeMode: ThemeMode,
    onThemeModeChange: (ThemeMode) -> Unit
) {
    var menuExpanded by remember { mutableStateOf(false) }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(10.dp))
            .clickable { menuExpanded = true }
            .padding(horizontal = 20.dp, vertical = 14.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(38.dp)
                .clip(CircleShape)
                .background(MaterialTheme.colorScheme.primaryContainer),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = Icons.Outlined.DarkMode,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onPrimaryContainer,
                modifier = Modifier.size(20.dp)
            )
        }
        Spacer(modifier = Modifier.width(14.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = "Dark Mode",
                fontSize = 15.sp,
                fontWeight = FontWeight.Medium,
                color = MaterialTheme.colorScheme.onSurface
            )
            Text(
                text = themeMode.label,
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
        Box {
            Icon(
                imageVector = Icons.Default.ArrowDropDown,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
            DropdownMenu(
                expanded = menuExpanded,
                onDismissRequest = { menuExpanded = false }
            ) {
                ThemeMode.entries.forEach { mode ->
                    DropdownMenuItem(
                        text = { Text(mode.label) },
                        trailingIcon = {
                            if (mode == themeMode) {
                                Icon(
                                    imageVector = Icons.Default.Check,
                                    contentDescription = null,
                                    tint = MaterialTheme.colorScheme.primary
                                )
                            }
                        },
                        onClick = {
                            menuExpanded = false
                            onThemeModeChange(mode)
                        }
                    )
                }
            }
        }
    }
}

@Composable
private fun SettingsAvatarButton(
    avatarUrl: String?,
    avatarBytes: ByteArray?,
    userName: String?,
    onClick: () -> Unit
) {
    val model = avatarBytes ?: avatarUrl
    Box(
        modifier = Modifier
            .padding(end = 16.dp)
            .size(40.dp)
            .clip(CircleShape)
            .background(MaterialTheme.colorScheme.primaryContainer)
            .clickable { onClick() },
        contentAlignment = Alignment.Center
    ) {
        if (model != null) {
            AsyncImage(
                model = model,
                contentDescription = "Profile",
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize()
            )
        } else {
            val initials = userName?.toInitials() ?: ""
            if (initials.isNotEmpty()) {
                Text(
                    text = initials,
                    color = MaterialTheme.colorScheme.onPrimaryContainer,
                    fontWeight = FontWeight.Bold,
                    fontSize = 16.sp
                )
            } else {
                Icon(
                    imageVector = Icons.Default.Person,
                    contentDescription = "Profile",
                    tint = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}