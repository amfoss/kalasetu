package com.example.kalasetu.features.feed

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Storefront
import androidx.compose.material.icons.outlined.Person
import androidx.compose.material3.Icon
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
    onStoreClick: () -> Unit,
    onHomeClick: () -> Unit,
    onProfileClick: () -> Unit
) {
    val items = listOf(
        Icons.Filled.Storefront,
        Icons.Filled.Home,
        Icons.Outlined.Person
    )

    NavigationBar(containerColor = Color.White) {
        items.forEachIndexed { index, icon ->
            NavigationBarItem(
                selected = selectedIndex == index,
                onClick = {
                    when (index) {
                        0 -> onStoreClick()
                        1 -> onHomeClick()
                        2 -> onProfileClick()
                    }
                },
                icon = { Icon(icon, contentDescription = null) },
                colors = NavigationBarItemDefaults.colors(
                    selectedIconColor = Color.White,
                    indicatorColor = SelectedPurple,
                    unselectedIconColor = SubtitleGray
                )
            )
        }
    }
}
