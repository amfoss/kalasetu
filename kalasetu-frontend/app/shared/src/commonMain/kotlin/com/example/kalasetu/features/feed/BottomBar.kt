package com.example.kalasetu.features.feed

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Storefront
import androidx.compose.material.icons.outlined.Event
import androidx.compose.material.icons.outlined.Person
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import com.example.kalasetu.theme.SelectedPurple
import com.example.kalasetu.theme.SubtitleGray
@Composable
internal fun KalaBottomNav(
    selectedIndex: Int,
    onHomeClick: () -> Unit,
    onEventsClick: () -> Unit,
    onStoreClick: () -> Unit,
    onProfileClick: () -> Unit
) {
    val items = listOf(
        Icons.Filled.Home,
        Icons.Outlined.Event,
        Icons.Filled.Storefront,
        Icons.Outlined.Person
    )

    NavigationBar(containerColor = MaterialTheme.colorScheme.surface) {
        items.forEachIndexed { index, icon ->
            NavigationBarItem(
                selected = selectedIndex == index,
                onClick = {
                    when (index) {
                        0 -> onHomeClick()
                        1 -> onEventsClick()
                        2 -> onStoreClick()
                        3 -> onProfileClick()
                    }
                },
                icon = {
                    Icon(
                        icon,
                        contentDescription = null
                    )
                },
                colors = NavigationBarItemDefaults.colors(
                    selectedIconColor = Color.White,
                    indicatorColor = SelectedPurple,
                    unselectedIconColor = SubtitleGray
                )
            )
        }
    }
}