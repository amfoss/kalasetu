package com.example.kalasetu.features.event

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.CalendarToday
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.foundation.background
import androidx.compose.material.icons.filled.Menu


@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OrganizerHomeScreen(
    userId: String,
    viewModel: EventListViewModel,
    onCreateEvent: () -> Unit,
    onEventClick: (String) -> Unit,
    onMenuClick: () -> Unit
) {
    val events by viewModel.events.collectAsStateWithLifecycle()
    LaunchedEffect(Unit) {
        viewModel.loadEvents(isOrganizer = true)
    }
    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Text(
                        "My Events",
                        fontWeight = FontWeight.Bold
                    )
                },
                navigationIcon = {
                    IconButton(
                        onClick = onMenuClick
                    ) {
                        Icon(
                            imageVector = Icons.Default.Menu,
                            contentDescription = "Open menu"
                        )
                    }
                },

                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = Color.White
                )
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = onCreateEvent,
                containerColor = PurplePrimary,
                contentColor = Color.White,
                shape = CircleShape,
                modifier = Modifier.size(56.dp),
            ) {
                Icon(
                    Icons.Default.Add,
                    contentDescription = "New Event",
                    modifier = Modifier.size(28.dp),
                )
            }
        },
        containerColor = Color.White,
    ) { padding ->
        if (events.isEmpty()) {
            // ─── Empty State ───
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
            ) {
                Text(
                    "No events yet",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark,
                )
                Spacer(Modifier.height(8.dp))
                Text(
                    "Tap + below to create your first event",
                    fontSize = 14.sp,
                    color = TextGray,
                )
            }
        } else {
            // ─── Event List ───
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
                contentPadding = PaddingValues(top = 8.dp, bottom = 100.dp),
            ) {
                items(events, key = { it.id }) { event ->
                    OrganizerEventCard(
                        event = event,
                        onClick = { onEventClick(event.id) },
                    )
                }
            }
        }
    }
}

@Composable
private fun OrganizerEventCard(event: Event, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, Color(0xFFE0E0E0)),
        elevation = CardDefaults.cardElevation(2.dp),
    ) {
        Column {
            // Cover image
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(150.dp),
            ) {
                if (event.coverImageBytes != null) {
                    coil3.compose.AsyncImage(
                        model = event.coverImageBytes,
                        contentDescription = event.title,
                        contentScale = androidx.compose.ui.layout.ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(Color(0xFFF4F1FF)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            event.title.take(2).uppercase(),
                            fontSize = 32.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFF7466F1),
                        )
                    }
                }
            }

            Column(modifier = Modifier.padding(16.dp)) {
                Text(event.title, fontSize = 18.sp, fontWeight = FontWeight.Bold, color = TextDark)
                Spacer(Modifier.height(6.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.LocationOn, null, tint = PurplePrimary, modifier = Modifier.size(16.dp))
                    Spacer(Modifier.width(4.dp))
                    Text(event.location, fontSize = 13.sp, color = TextGray)
                }
                Spacer(Modifier.height(4.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.CalendarToday, null, tint = PurplePrimary, modifier = Modifier.size(16.dp))
                    Spacer(Modifier.width(4.dp))
                    Text(formatEventDuration(event.startDate, event.endDate), fontSize = 13.sp, color = TextGray)
                }
                Spacer(Modifier.height(10.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        "View Applications",
                        fontSize = 12.sp,
                        color = PurplePrimary,
                        fontWeight = FontWeight.Medium,
                    )
                    Spacer(Modifier.width(4.dp))
                    Icon(
                        Icons.AutoMirrored.Filled.ArrowForward,
                        null,
                        tint = PurplePrimary,
                        modifier = Modifier.size(14.dp),
                    )
                }
            }
        }
    }
}