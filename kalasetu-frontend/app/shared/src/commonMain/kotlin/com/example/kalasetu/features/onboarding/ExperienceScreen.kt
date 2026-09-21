package com.example.kalasetu.features.onboarding

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
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

@Composable
fun ExperienceScreen(
    onNext: (String) -> Unit,
    onBack: () -> Unit,
){
    var selectedLevel by remember { mutableStateOf("") }
    val levels = listOf("Just Starting Out", "Learning & Growing", "Experienced", "Professional")

    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
    ) {
        Column(
            modifier = Modifier.fillMaxSize()
        ) {
            IconButton(onClick = onBack) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Where Are You In Your\nJourney?",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Personalized for you.",
                fontSize = 24.sp,
                color = SubtitleGray,
            )

            Spacer(modifier = Modifier.height(32.dp))

            Text(
                text = "Experience Level",
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
            )

            Spacer(modifier = Modifier.height(12.dp))

            levels.forEach { level ->
                OutlinedButton(
                    onClick = { selectedLevel = level },
                    shape = RoundedCornerShape(25.dp),
                    border = BorderStroke(
                        1.dp,
                        if (selectedLevel == level) SelectedPurple
                        else UnselectedBorder
                    ),
                    colors = ButtonDefaults.outlinedButtonColors(
                        containerColor = if (selectedLevel == level)
                            SelectedPurple.copy(alpha = 0.08f)
                        else Color.Transparent,
                        contentColor = Color(0xFF1C1B1F),
                    ),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(80.dp)
                        .padding(vertical = 11.dp),
                ) {
                    Text(level)
                }
            }
        }

        IconButton(
            onClick = { onNext(selectedLevel) },
            enabled = selectedLevel.isNotBlank(),
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .size(56.dp)
                .clip(RoundedCornerShape(28.dp))
                .background(
                    if (selectedLevel.isNotBlank()) MaterialTheme.colorScheme.primary
                    else MaterialTheme.colorScheme.surfaceVariant
                ),
        ) {
            Icon(
                imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                contentDescription = "Next",
                tint = if (selectedLevel.isNotBlank()) MaterialTheme.colorScheme.onPrimary
                       else MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}
