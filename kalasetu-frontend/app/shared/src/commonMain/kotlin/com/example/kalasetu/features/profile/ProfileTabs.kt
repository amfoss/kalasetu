package com.example.kalasetu.features.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun ProfileTabBar(
    selected: ProfileTab,
    onTabSelected: (ProfileTab) -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        ProfileTab.entries.forEach { tab ->

            val isSelected = tab == selected

            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(20.dp))
                    .background(
                        if (isSelected) BrandPurple
                        else Color.Transparent
                    )
                    .border(
                        width = if (isSelected) 0.dp else 1.dp,
                        color = if (isSelected) {
                            Color.Transparent
                        } else {
                            MaterialTheme.colorScheme.outlineVariant
                        },
                        shape = RoundedCornerShape(20.dp)
                    )
                    .clickable {
                        onTabSelected(tab)
                    }
                    .padding(
                        horizontal = 18.dp,
                        vertical = 8.dp
                    )
            ) {
                Text(
                    text = tab.name
                        .lowercase()
                        .replaceFirstChar { it.uppercase() },
                    fontSize = 14.sp,
                    fontWeight = if (isSelected) {
                        FontWeight.SemiBold
                    } else {
                        FontWeight.Normal
                    },
                    color = if (isSelected) {
                        Color.White
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    }
                )
            }
        }
    }
}