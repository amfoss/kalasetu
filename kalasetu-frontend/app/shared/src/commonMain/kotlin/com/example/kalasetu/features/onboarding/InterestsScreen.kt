package com.example.kalasetu.features.onboarding

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.theme.SelectedPurple
import com.example.kalasetu.theme.SubtitleGray
import com.example.kalasetu.theme.UnselectedBorder

private data class InterestSection(val title: String, val options: List<String>)

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun InterestsScreen(
    onNext: (List<String>) -> Unit,
    onBack: () -> Unit,
    ) {
    val selectedPills = remember { mutableStateMapOf<String, Boolean>() }
    val scrollState = rememberScrollState()

    val sections = listOf(
        InterestSection("Performing Arts", listOf("Music", "Dance", "Theatre", "Standup-comedy")),
        InterestSection("Visual & Creative Arts", listOf("Painting", "Photography", "Digital Art", "Graphic Design")),
        InterestSection("Events and Experiences", listOf("Live Shows", "Workshops", "Exhibitions", "Festivals")),
        InterestSection("Social & Community", listOf("Meetups", "Networking", "Local Communities", "Cultural Events")),
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(scrollState),
        ) {
            IconButton(onClick = onBack) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "What Are You Interested In?",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Pick interests to see what you'll love.",
                fontSize = 24.sp,
                color = SubtitleGray,
            )

            Spacer(modifier = Modifier.height(32.dp))

            sections.forEach { section ->
                Text(
                    text = section.title,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                )

                Spacer(modifier = Modifier.height(12.dp))

                FlowRow(
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    section.options.forEach { pill ->
                        val isSelected = selectedPills[pill] ?: false
                        OutlinedButton(
                            onClick = {
                                selectedPills[pill] = !isSelected
                            },
                            shape = RoundedCornerShape(25.dp),
                            border = BorderStroke(
                                1.dp,
                                if (isSelected) SelectedPurple
                                else UnselectedBorder
                            ),
                            colors = ButtonDefaults.outlinedButtonColors(
                                containerColor = if (isSelected)
                                    SelectedPurple.copy(alpha = 0.1f)
                                else Color.Transparent,
                                contentColor = if (isSelected) SelectedPurple
                                else MaterialTheme.colorScheme.onSurfaceVariant,
                            ),
                            contentPadding = PaddingValues(horizontal = 16.dp, vertical = 6.dp),
                        ) {
                            Text(pill)
                        }
                    }
                }

                Spacer(modifier = Modifier.height(24.dp))
            }

            Spacer(modifier = Modifier.height(80.dp))
        }

        IconButton(
            onClick = {
                val selectedInterests = selectedPills
                    .filter { it.value }
                    .keys
                    .toList()

                onNext(selectedInterests)
            },
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .size(56.dp)
                .clip(RoundedCornerShape(28.dp))
                .background(MaterialTheme.colorScheme.primary),
        ) {
            Icon(
                imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                contentDescription = "Next",
                tint = MaterialTheme.colorScheme.onPrimary,
            )
        }
    }
}
