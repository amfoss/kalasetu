package com.example.kalasetu.features.event

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@OptIn(ExperimentalLayoutApi::class, ExperimentalMaterial3Api::class)
@Composable
fun SelectArtistCategoriesScreen(
    onNext: (List<String>) -> Unit,
    onBack: () -> Unit, 
) {
    val selectedCategories = remember { mutableStateListOf<String>() }

    fun toggleCategory(category: String) {
        if (selectedCategories.contains(category)) {
            selectedCategories.remove(category)
        } else {
            selectedCategories.add(category)
        }
    }

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(
                onClick = { onNext(selectedCategories.toList()) },
                containerColor = LightPurpleBg,
                contentColor = TextDark,
                shape = CircleShape,
                modifier = Modifier.size(56.dp),
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                    contentDescription = "Next",
                    modifier = Modifier.size(24.dp),
                )
            }
        },
        containerColor = MaterialTheme.colorScheme.surface,
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(horizontal = 24.dp)
                .verticalScroll(rememberScrollState())
        ) {
            // Adjusted top padding since there is no TopAppBar
            Spacer(modifier = Modifier.height(40.dp))
            IconButton(
                onClick = onBack,
                modifier = Modifier.offset(x = (-12).dp) // Slight offset to align visually with the edge
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                    tint = TextDark
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            // ─── TITLE ───
            Text(
                text = "Select Artist Categories",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark,
                lineHeight = 38.sp
            )

            Spacer(modifier = Modifier.height(12.dp))

            // ─── SUBTITLE (Exact font from design) ───
            Text(
                text = "Pick the art you want to\nexperience",
                fontSize = 24.sp,
                fontWeight = FontWeight.Normal, // Lighter weight
                color = TextGray, // Faded gray color
                lineHeight = 32.sp
            )

            Spacer(modifier = Modifier.height(40.dp))

            // ─── CATEGORIES SECTIONS ───
            CategorySection(
                title = "Performing Arts",
                categories = listOf("Music", "Dance", "Theatre", "Stand-up Comedy"),
                selectedCategories = selectedCategories,
                onToggle = ::toggleCategory
            )

            CategorySection(
                title = "Visual & Creative Arts",
                categories = listOf("Painting", "Photography", "Digital Art", "Graphic Design"),
                selectedCategories = selectedCategories,
                onToggle = ::toggleCategory
            )

            CategorySection(
                title = "Events & Experiences",
                categories = listOf("Live Shows", "Workshops", "Exhibitions", "Festivals"),
                selectedCategories = selectedCategories,
                onToggle = ::toggleCategory
            )

            CategorySection(
                title = "Social & Community",
                categories = listOf("Meetups", "Networking", "Local Communities", "Cultural Events"),
                selectedCategories = selectedCategories,
                onToggle = ::toggleCategory
            )

            Spacer(modifier = Modifier.height(80.dp)) // Space for FAB
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun CategorySection(
    title: String,
    categories: List<String>,
    selectedCategories: List<String>,
    onToggle: (String) -> Unit
) {
    Column(modifier = Modifier.padding(bottom = 28.dp)) {
        Text(
            text = title,
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            color = TextDark,
            modifier = Modifier.padding(bottom = 12.dp)
        )
        FlowRow(
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            categories.forEach { category ->
                val isSelected = selectedCategories.contains(category)
                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(50)) // Fully rounded pill shape
                        .background(if (isSelected) MaterialTheme.colorScheme.primaryContainer else MaterialTheme.colorScheme.surface)
                        .border(
                            width = 1.dp,
                            color = if (isSelected) PurplePrimary else BorderGray,
                            shape = RoundedCornerShape(50)
                        )
                        .clickable { onToggle(category) }
                        .padding(horizontal = 16.dp, vertical = 8.dp)
                ) {
                    Text(
                        text = category,
                        fontSize = 14.sp,
                        color = if (isSelected) PurplePrimary else TextGray,
                        fontWeight = if (isSelected) FontWeight.Medium else FontWeight.Normal
                    )
                }
            }
        }
    }
}