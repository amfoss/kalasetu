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
fun OrganizerTypeScreen(
    onNext: () -> Unit,
    onBack: () -> Unit,
) {
    var selectedType by remember { mutableStateOf("") }
    val types = listOf("Event Organizer", "Brand / Company", "Community / Club", "Individual / Freelancer")

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
                text = "What Kind Of Organizer\nAre You?",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Connect with the right talent.",
                fontSize = 24.sp,
                color = SubtitleGray,
            )

            Spacer(modifier = Modifier.height(32.dp))

            Text(
                text = "Organizer Type",
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
            )

            Spacer(modifier = Modifier.height(12.dp))

            types.forEach { type ->
                OutlinedButton(
                    onClick = { selectedType = type },
                    shape = RoundedCornerShape(25.dp),
                    border = BorderStroke(
                        1.dp,
                        if (selectedType == type) SelectedPurple
                        else UnselectedBorder
                    ),
                    colors = ButtonDefaults.outlinedButtonColors(
                        containerColor = if (selectedType == type)
                            SelectedPurple.copy(alpha = 0.08f)
                        else Color.Transparent,
                        contentColor = MaterialTheme.colorScheme.onSurface,
                    ),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(80.dp)
                        .padding(vertical = 11.dp),
                ) {
                    Text(type)
                }
            }
        }

        IconButton(
            onClick = onNext,
            enabled = selectedType.isNotBlank(),
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .size(56.dp)
                .clip(RoundedCornerShape(28.dp))
                .background(
                    if (selectedType.isNotBlank()) MaterialTheme.colorScheme.primary
                    else MaterialTheme.colorScheme.surfaceVariant
                ),
        ) {
            Icon(
                imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                contentDescription = "Next",
                tint = if (selectedType.isNotBlank()) MaterialTheme.colorScheme.onPrimary
                       else MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}
