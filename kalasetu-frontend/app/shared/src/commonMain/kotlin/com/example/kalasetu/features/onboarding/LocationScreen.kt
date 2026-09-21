package com.example.kalasetu.features.onboarding

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.theme.SubtitleGray
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OnboardingLocationScreen(
    onNext: (String) -> Unit,
    onBack: () -> Unit,
) {
    var query by remember { mutableStateOf("") }
    var expanded by remember { mutableStateOf(false) }

    val selectedCountry = Countries.firstOrNull { it.equals(query, ignoreCase = true) }
    val filteredCountries = if (query.isBlank()) {
        Countries
    } else {
        Countries.filter { it.contains(query, ignoreCase = true) }
    }

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
                text = "Where are you based?",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "To personalize your\nexperience.",
                fontSize = 24.sp,
                color = SubtitleGray
            )

            Spacer(modifier = Modifier.height(32.dp))

            Text(
                text = "Country",
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
            )

            Spacer(modifier = Modifier.height(8.dp))

//            ExposedDropdownMenuBox(
//                expanded = expanded,
//                onExpandedChange = { expanded = it },
//            ) {
            TextField(
                value = query,
                onValueChange = {
                    query = it
                },
                placeholder = { Text("Enter your Country") },
                leadingIcon = {
                    Icon(
                        imageVector = Icons.Filled.LocationOn,
                        contentDescription = null,
                    )
                },
                singleLine = true,
                shape = RoundedCornerShape(10.dp),
                modifier = Modifier.fillMaxWidth()
            )

//                ExposedDropdownMenu(
//                    expanded = expanded,
//                    onDismissRequest = { expanded = false },
//                    modifier = Modifier.height(300.dp),
//                ) {
//                    filteredCountries.forEach { country ->
//                        DropdownMenuItem(
//                            text = { Text(country) },
//                            onClick = {
//                                query = country
//                                expanded = false
//                            },
//                        )
//                    }
//                }
            }


        IconButton(
            onClick = { selectedCountry?.let { onNext(it) } },
            enabled = selectedCountry != null,
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .size(56.dp)
                .clip(RoundedCornerShape(28.dp))
                .background(
                    if (selectedCountry != null) MaterialTheme.colorScheme.primary
                    else MaterialTheme.colorScheme.surfaceVariant
                ),
        ) {
            Icon(
                imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                contentDescription = "Next",
                tint = if (selectedCountry != null) MaterialTheme.colorScheme.onPrimary
                       else MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }}
