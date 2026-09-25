package com.example.kalasetu.features.event

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CalendarToday
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.Email
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Phone
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil3.compose.AsyncImage
import kotlinx.datetime.LocalDate

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun ReviewEventScreen(
    draft: EventDraft,
    onBack: () -> Unit,
    onEdit: () -> Unit,
    onPublish: () -> Unit,
) {
    Scaffold(containerColor = MaterialTheme.colorScheme.surface) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 24.dp),
        ) {
            Spacer(Modifier.height(16.dp))

            // ─── Back Arrow ───
            IconButton(
                onClick = onBack,
                modifier = Modifier.offset(x = (-12).dp)
            ) {
                Icon(
                    Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                    tint = MaterialTheme.colorScheme.onSurface,
                )
            }

            Spacer(Modifier.height(8.dp))

            // ─── Title ───
            Text(
                text = "Review your event",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark,
                lineHeight = 38.sp,
            )

            Spacer(Modifier.height(8.dp))

            // ─── Subtitle ───
            Text(
                text = "Confirm the details and entertain the audience",
                fontSize = 18.sp,
                color = TextGray,
                lineHeight = 24.sp,
            )

            Spacer(Modifier.height(28.dp))

            // ─── EVENT CARD ───
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                border = BorderStroke(1.dp, BorderGray),
                elevation = CardDefaults.cardElevation(2.dp),
            ) {
                Column {
                    // ─── Cover Image (organizer badge removed) ───
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(220.dp),
                    ) {
                        if (draft.coverImageBytes != null) {
                            AsyncImage(
                                model = draft.coverImageBytes,
                                contentDescription = "Event Cover",
                                contentScale = ContentScale.Crop,
                                modifier = Modifier.fillMaxSize(),
                            )
                        } else {
                            Box(
                                modifier = Modifier
                                    .fillMaxSize()
                                    .background(LightPurpleBg),
                                contentAlignment = Alignment.Center,
                            ) {
                                Text(
                                    "No Cover Image",
                                    color = PurplePrimary,
                                    fontWeight = FontWeight.Bold,
                                )
                            }
                        }
                    }

                    // ─── Details Section ───
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = draft.title.ifBlank { "Event title" },
                            fontSize = 20.sp,
                            fontWeight = FontWeight.Bold,
                            color = TextDark,
                        )
                        Spacer(Modifier.height(4.dp))
                        Text(
                            text = draft.description.ifBlank { "The event description" },
                            fontSize = 13.sp,
                            color = TextGray,
                            lineHeight = 18.sp,
                        )

                        Spacer(Modifier.height(16.dp))

                        // ─── CATEGORIES ───
                        if (draft.categories.isNotEmpty()) {
                            Text(
                                text = "Looking for",
                                fontSize = 13.sp,
                                fontWeight = FontWeight.Bold,
                                color = TextDark,
                            )
                            Spacer(Modifier.height(8.dp))
                            FlowRow(
                                horizontalArrangement = Arrangement.spacedBy(6.dp),
                                verticalArrangement = Arrangement.spacedBy(6.dp),
                            ) {
                                draft.categories.forEach { category ->
                                    Box(
                                        modifier = Modifier
                                            .clip(RoundedCornerShape(50))
                                            .background(LightPurpleBg)
                                            .border(1.dp, PurplePrimary, RoundedCornerShape(50))
                                            .padding(horizontal = 12.dp, vertical = 6.dp),
                                    ) {
                                        Text(
                                            text = category,
                                            fontSize = 12.sp,
                                            color = PurplePrimary,
                                            fontWeight = FontWeight.Medium,
                                        )
                                    }
                                }
                            }
                            Spacer(Modifier.height(16.dp))
                        }

                        // ─── Detail Rows ───
                        DetailRow(
                            icon = Icons.Default.CalendarToday,
                            label = "Duration",
                            value = formatDateRange(draft.startDate, draft.endDate),
                        )
                        DetailRow(
                            icon = Icons.Default.Email,
                            label = "Email",
                            value = draft.email.ifBlank { "Not provided" },
                        )
                        DetailRow(
                            icon = Icons.Default.Phone,
                            label = "Phone",
                            value = draft.phone.ifBlank { "Not provided" },
                        )
                        DetailRow(
                            icon = Icons.Default.LocationOn,
                            label = "Location",
                            value = draft.location.ifBlank { "Not provided" },
                        )
                    }
                }
            }

            Spacer(Modifier.height(24.dp))

            // ─── Info Banner ───
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(RoundedCornerShape(8.dp))
                    .background(LightPurpleBg)
                    .padding(14.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(
                    Icons.Default.Info,
                    contentDescription = null,
                    tint = PurplePrimary,
                    modifier = Modifier.size(20.dp),
                )
                Spacer(Modifier.width(12.dp))
                Text(
                    text = "Please review all the details carefully before publishing",
                    fontSize = 13.sp,
                    color = TextDark,
                    lineHeight = 18.sp,
                )
            }

            Spacer(Modifier.height(20.dp))

            // ─── Action Buttons ───
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                OutlinedButton(
                    onClick = onEdit,
                    modifier = Modifier.weight(1f).height(52.dp),
                    shape = RoundedCornerShape(12.dp),
                    border = BorderStroke(1.5.dp, PurplePrimary),
                    colors = ButtonDefaults.outlinedButtonColors(contentColor = PurplePrimary),
                ) {
                    Icon(Icons.Default.Edit, contentDescription = null, modifier = Modifier.size(18.dp))
                    Spacer(Modifier.width(8.dp))
                    Text("Edit", fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }

                Button(
                    onClick = onPublish,
                    modifier = Modifier.weight(1f).height(52.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = PurplePrimary),
                ) {
                    Icon(Icons.Default.Send, contentDescription = null, modifier = Modifier.size(18.dp))
                    Spacer(Modifier.width(8.dp))
                    Text("Publish", fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }
            }

            Spacer(Modifier.height(40.dp))
        }
    }
}

@Composable
private fun DetailRow(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    value: String,
) {
    Row(
        modifier = Modifier.fillMaxWidth().padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .clip(RoundedCornerShape(8.dp))
                .background(LightPurpleBg),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = null, tint = PurplePrimary, modifier = Modifier.size(18.dp))
        }
        Spacer(Modifier.width(12.dp))
        Text(
            text = label,
            fontSize = 13.sp,
            color = TextGray,
            modifier = Modifier.width(72.dp),
        )
        Text(
            text = value,
            fontSize = 13.sp,
            fontWeight = FontWeight.Medium,
            color = TextDark,
            modifier = Modifier.weight(1f),
        )
    }
}

// ─── Formats dates like: "9 sep 2025 to 13 sep 2025" ───
fun formatDateRange(start: LocalDate?, end: LocalDate?): String {
    if (start == null) return "Not selected"
    val months = listOf(
        "jan", "feb", "mar", "apr", "may", "jun",
        "jul", "aug", "sep", "oct", "nov", "dec"
    )
    // ✅ Now includes the year
    val startStr = "${start.dayOfMonth} ${months[start.monthNumber - 1]} ${start.year}"

    if (end == null || end == start) return startStr

    // If same month & year, keep it short: "9 to 13 sep 2025"
    val endStr = if (end.monthNumber == start.monthNumber && end.year == start.year) {
        "${end.dayOfMonth}"
    } else {
        "${end.dayOfMonth} ${months[end.monthNumber - 1]} ${end.year}"
    }

    return if (end.monthNumber == start.monthNumber && end.year == start.year) {
        "${start.dayOfMonth} to $endStr ${months[start.monthNumber - 1]} ${start.year}"
    } else {
        "$startStr to $endStr"
    }
}