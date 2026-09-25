package com.example.kalasetu.features.application

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

private val PurplePrimary = Color(0xFF7466F1)
private val LightPurpleBg @Composable get() = MaterialTheme.colorScheme.primaryContainer
private val TextDark @Composable get() = MaterialTheme.colorScheme.onSurface
private val TextGray @Composable get() = MaterialTheme.colorScheme.onSurfaceVariant
private val BorderGray @Composable get() = MaterialTheme.colorScheme.outlineVariant
private val SuccessGreen = Color(0xFF4CAF50)
private val SuccessGreenBg = Color(0xFFE8F5E9)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ApplicationStatusScreen(
    eventTitle: String,
    onBack: () -> Unit,
    onDone: () -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        "Application Status",
                        fontWeight = FontWeight.Bold,
                        fontSize = 20.sp,
                    )
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(
                            Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "Back",
                            tint = PurplePrimary,
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.surface),
            )
        },
        containerColor = MaterialTheme.colorScheme.surface,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 24.dp),
        ) {
            Spacer(Modifier.height(8.dp))

            // ─── Main Status Card ───
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                border = androidx.compose.foundation.BorderStroke(1.dp, BorderGray),
                elevation = CardDefaults.cardElevation(2.dp),
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    // Clock icon in white circle
                    Box(
                        modifier = Modifier
                            .size(90.dp)
                            .clip(CircleShape)
                            .background(MaterialTheme.colorScheme.surface),
                        contentAlignment = Alignment.Center,
                    ) {
                        Box(
                            modifier = Modifier
                                .size(72.dp)
                                .clip(CircleShape)
                                .border(3.dp, PurplePrimary, CircleShape),
                            contentAlignment = Alignment.Center,
                        ) {
                            Icon(
                                Icons.Default.Schedule,
                                contentDescription = null,
                                tint = PurplePrimary,
                                modifier = Modifier.size(40.dp),
                            )
                        }
                    }

                    Spacer(Modifier.height(20.dp))

                    Text(
                        "Under review",
                        fontSize = 22.sp,
                        fontWeight = FontWeight.Bold,
                        color = PurplePrimary,
                    )

                    Spacer(Modifier.height(12.dp))

                    Text(
                        "Your application is currently under review. We will notify once the process is done",
                        fontSize = 13.sp,
                        color = TextDark,
                        textAlign = TextAlign.Center,
                        lineHeight = 19.sp,
                    )

                    Spacer(Modifier.height(16.dp))

                    // Purple info pill
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(50))
                            .background(LightPurpleBg)
                            .padding(horizontal = 14.dp, vertical = 8.dp),
                    ) {
                        Row(verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.Center, ) {
                            Icon(
                                Icons.Default.Schedule,
                                contentDescription = null,
                                tint = PurplePrimary,
                                modifier = Modifier.size(14.dp),
                            )
                            Spacer(Modifier.width(6.dp))
                            Text(
                                "Review will be done in ~3 days",
                                fontSize = 11.sp,
                                color = PurplePrimary,
                                fontWeight = FontWeight.Medium,
                                maxLines = 1,
                                softWrap = false,
                            )
                        }
                    }
                }
            }

            Spacer(Modifier.height(28.dp))

            // ─── Status Timeline Heading ───
            Text(
                "Status Timeline",
                fontSize = 20.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark,
            )

            Spacer(Modifier.height(16.dp))

            // ─── Timeline Items ───
            TimelineItem(
                icon = { Icon(Icons.Default.Check, null, tint = Color.White, modifier = Modifier.size(20.dp)) },
                iconBg = SuccessGreen,
                isDashed = false,
                showConnector = true,
                title = "Submitted",
                subtitle = "Your application is submitted",
                boxBg = SuccessGreenBg,
                boxBorder = SuccessGreen.copy(alpha = 0.4f),
            )
            TimelineItem(
                icon = { Icon(Icons.Default.Schedule, null, tint = Color.White, modifier = Modifier.size(20.dp)) },
                iconBg = PurplePrimary,
                isDashed = false,
                showConnector = true,
                title = "Review",
                subtitle = "Your application is being reviewed",
                boxBg = LightPurpleBg,
                boxBorder = PurplePrimary.copy(alpha = 0.4f),
            )
            TimelineItem(
                icon = { Icon(Icons.Default.Notifications, null, tint = TextGray, modifier = Modifier.size(20.dp)) },
                iconBg = Color.Transparent,
                isDashed = true,
                showConnector = false,
                title = "Notification",
                subtitle = "You will be notified when the process is done",
                boxBg = MaterialTheme.colorScheme.surface,
                boxBorder = BorderGray,
            )

            Spacer(Modifier.height(32.dp))

            // ─── Done Button ───
            Button(
                onClick = onDone,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(containerColor = PurplePrimary),
            ) {
                Text("Done", fontSize = 16.sp, fontWeight = FontWeight.Bold)
            }

            Spacer(Modifier.height(32.dp))
        }
    }
}

@Composable
private fun TimelineItem(
    icon: @Composable () -> Unit,
    iconBg: Color,
    isDashed: Boolean,
    showConnector: Boolean,
    title: String,
    subtitle: String,
    boxBg: Color,
    boxBorder: Color,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(IntrinsicSize.Min),
    ) {
        // ─── Left: Icon + Connector ───
        Column(
            modifier = Modifier.width(48.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Box(
                modifier = Modifier
                    .size(44.dp)
                    .clip(CircleShape)
                    .then(
                        if (isDashed) {
                            Modifier.border(1.5.dp, BorderGray, CircleShape)
                        } else {
                            Modifier.background(iconBg)
                        }
                    ),
                contentAlignment = Alignment.Center,
            ) {
                icon()
            }

            if (showConnector) {
                Box(
                    modifier = Modifier
                        .width(2.dp)
                        .height(50.dp)
                        .background(PurplePrimary.copy(alpha = 0.5f)),
                )
            }
        }

        Spacer(Modifier.width(12.dp))

        // ─── Right: Content Box ───
        Box(
            modifier = Modifier
                .weight(1f)
                .padding(bottom = if (showConnector) 12.dp else 0.dp)
                .clip(RoundedCornerShape(12.dp))
                .background(boxBg)
                .border(1.dp, boxBorder, RoundedCornerShape(12.dp))
                .padding(horizontal = 14.dp, vertical = 14.dp),
        ) {
            Column {
                Text(
                    title,
                    fontSize = 15.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark,
                )
                Spacer(Modifier.height(4.dp))
                Text(
                    subtitle,
                    fontSize = 12.sp,
                    color = TextGray,
                    lineHeight = 16.sp,
                )
            }
        }
    }
}