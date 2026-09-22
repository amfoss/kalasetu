package com.example.kalasetu.features.event

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
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
import com.example.kalasetu.features.feed.KalaBottomNav

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ArtistHomeScreen(
    viewModel: EventListViewModel,
    onEventClick: (String) -> Unit,
    onSwitchRole: () -> Unit = {},
    onMenuClick: () -> Unit,
    onMyEventsClick: () -> Unit,
    onStoreClick: () -> Unit,
    onEventsClick: () -> Unit,
    onHomeClick: () -> Unit,
    onProfileClick: () -> Unit
) {
    val events by viewModel.events.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.loadEvents(isOrganizer = false)
    }

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Text(
                        text = "Discover Events",
                        fontWeight = FontWeight.Bold
                    )
                },
                navigationIcon = {
                    IconButton(onClick = onMenuClick) {
                        Icon(
                            Icons.Default.Menu,
                            contentDescription = "Menu"
                        )
                    }
                },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = Color.White
                )
            )
        },

        // ─── Bottom Navigation ───
        bottomBar = {
            KalaBottomNav(
                selectedIndex = 1, // Events
                onStoreClick = onStoreClick,
                onEventsClick = onEventsClick,
                onHomeClick = onHomeClick,
                onProfileClick = onProfileClick
            )
        },

        containerColor = Color.White
    ) { padding ->

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {

            Button(
                onClick = onMyEventsClick,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                shape = RoundedCornerShape(10.dp)
            ) {
                Text(
                    text = "My Events",
                    fontWeight = FontWeight.SemiBold
                )
            }

            if (events.isEmpty()) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        "No events yet",
                        color = TextGray
                    )
                }
            } else {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(1f)
                        .padding(horizontal = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    item {
                        Spacer(Modifier.height(8.dp))
                    }

                    items(events) { event ->
                        EventCard(event = event) {
                            onEventClick(event.id)
                        }
                    }

                    item {
                        Spacer(Modifier.height(16.dp))
                    }
                }
            }
        }
    }
}

@Composable
fun EventCard(event: Event, onClick: () -> Unit) {
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
            // ─── Cover image ───
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
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

            // ─── Details ───
            Column(modifier = Modifier.padding(16.dp)) {
                Text(
                    event.title,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF1E1E1E),
                )
                Spacer(Modifier.height(8.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.LocationOn,
                        contentDescription = null,
                        tint = Color(0xFF7466F1),
                        modifier = Modifier.size(16.dp),
                    )
                    Spacer(Modifier.width(4.dp))
                    Text(event.location, fontSize = 13.sp, color = Color(0xFF757575))
                }
                Spacer(Modifier.height(4.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.CalendarToday,
                        contentDescription = null,
                        tint = Color(0xFF7466F1),
                        modifier = Modifier.size(16.dp),
                    )
                    Spacer(Modifier.width(4.dp))
                    Text(
                        formatEventDuration(event.startDate, event.endDate),
                        fontSize = 13.sp,
                        color = Color(0xFF757575),
                    )
                }
                Spacer(Modifier.height(8.dp))
                Text(
                    "By ${event.organizerName}",
                    fontSize = 12.sp,
                    color = Color(0xFF757575),
                )
            }
        }
    }
}