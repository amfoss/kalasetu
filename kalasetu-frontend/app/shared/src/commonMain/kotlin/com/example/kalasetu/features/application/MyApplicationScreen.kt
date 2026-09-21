package com.example.kalasetu.features.application

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.KeyboardArrowRight
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Image
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil3.compose.AsyncImage
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
private val PurplePrimary = Color(0xFF7466F1)
private val LightPurpleBg = Color(0xFFF4F1FF)
private val TextDark = Color(0xFF1E1E1E)
private val TextGray = Color(0xFF757575)
private val BorderGray = Color(0xFFE0E0E0)

private val StatusPendingBg = Color(0xFFFFF4E5)
private val StatusPendingFg = Color(0xFFE69500)
private val StatusAcceptedBg = Color(0xFFE5F7E5)
private val StatusAcceptedFg = Color(0xFF2E7D32)
private val StatusRejectedBg = Color(0xFFFFE5E5)
private val StatusRejectedFg = Color(0xFFD32F2F)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MyApplicationsScreen(
    onApplicationClick: (String) -> Unit,
    onBack: () -> Unit,
    onSwitchRole: () -> Unit = {},
    onMenuClick: () -> Unit
){
    LaunchedEffect(Unit) {
        ApplicationRepository.fetchMyApplications()
    }

    val allApplications by ApplicationStore.applications.collectAsState()
    val tabs = listOf("All", "Pending", "Accepted", "Rejected")
    var selectedTab by remember { mutableStateOf(0) }

    val filtered = remember(allApplications, selectedTab) {
        when (selectedTab) {
            1 -> allApplications.filter { it.status == ApplicationStatus.PENDING }
            2 -> allApplications.filter { it.status == ApplicationStatus.ACCEPTED }
            3 -> allApplications.filter { it.status == ApplicationStatus.REJECTED }
            else -> allApplications
        }
    }

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Text(
                        text = "My Applications",
                        fontWeight = FontWeight.Bold
                    )
                },
                navigationIcon = {
                    IconButton(onClick = onMenuClick) {
                        IconButton(onClick = onMenuClick) {
                            Icon(Icons.Default.Menu, contentDescription = "Menu")
                        }
                    }
                },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = Color.White
                )
            )
        },
        containerColor = Color.White,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            // ─── Tabs ───
            ScrollableTabRow(
                selectedTabIndex = selectedTab,
                containerColor = Color.White,
                contentColor = PurplePrimary,
                edgePadding = 0.dp,
                divider = {},
                indicator = { tabPositions ->

                    Box(
                        modifier = Modifier
                            .tabIndicatorOffset(tabPositions[selectedTab])
                            .wrapContentSize(Alignment.Center)
                    ) {
                        Box(
                            modifier = Modifier
                                .width(35.dp)
                                .height(3.dp)
                                .clip(RoundedCornerShape(50))
                                .background(PurplePrimary)
                        )
                    }
                }
            ) {
                tabs.forEachIndexed { index, title ->
                    Tab(
                        selected = selectedTab == index,
                        onClick = {
                            selectedTab = index
                        },
                        text = {
                            Text(
                                text = title,
                                fontSize = 14.sp,
                                fontWeight = if (selectedTab == index) {
                                    FontWeight.Bold
                                } else {
                                    FontWeight.Normal
                                },
                                color = if (selectedTab == index) {
                                    PurplePrimary
                                } else {
                                    TextGray
                                }
                            )
                        }
                    )
                }
            }

            // Purple underline for active tab

            Spacer(Modifier.height(12.dp))

            if (filtered.isEmpty()) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text("No applications yet", color = TextGray)
                }
            } else {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    items(filtered, key = { it.id }) { app ->
                        ApplicationRow(app = app, onClick = { onApplicationClick(app.id) })
                    }
                    item { Spacer(Modifier.height(16.dp)) }
                }
            }
        }
    }
}

@Composable
private fun ApplicationRow(app: Application, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, BorderGray),
        elevation = CardDefaults.cardElevation(0.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            // Cover thumbnail
            Box(
                modifier = Modifier
                    .size(56.dp)
                    .clip(RoundedCornerShape(8.dp))
                    .background(LightPurpleBg),
                contentAlignment = Alignment.Center,
            ) {
                if (app.eventCoverBytes != null) {
                    AsyncImage(
                        model = app.eventCoverBytes,
                        contentDescription = null,
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    Icon(Icons.Default.Image, null, tint = PurplePrimary)
                }
            }

            Spacer(Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    app.eventTitle.ifBlank { "Untitled Event" },
                    fontSize = 15.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark,
                )
                Spacer(Modifier.height(4.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.Check,
                        contentDescription = null,
                        tint = TextGray,
                        modifier = Modifier.size(14.dp),
                    )
                    Spacer(Modifier.width(4.dp))
                    Text(
                        "Perfect will check it !",
                        fontSize = 12.sp,
                        color = TextGray,
                    )
                }
            }

            Spacer(Modifier.width(8.dp))

            // Status badge
            StatusBadge(status = app.status)

            Icon(
                Icons.AutoMirrored.Filled.KeyboardArrowRight,
                contentDescription = null,
                tint = TextGray,
                modifier = Modifier.size(20.dp),
            )
        }
    }
}

@Composable
private fun StatusBadge(status: ApplicationStatus) {
    val (bg, fg) = when (status) {
        ApplicationStatus.PENDING -> StatusPendingBg to StatusPendingFg
        ApplicationStatus.ACCEPTED -> StatusAcceptedBg to StatusAcceptedFg
        ApplicationStatus.REJECTED -> StatusRejectedBg to StatusRejectedFg
    }
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(50))
            .background(bg)
            .padding(horizontal = 10.dp, vertical = 4.dp),
    ) {
        Text(
            status.label,
            fontSize = 11.sp,
            fontWeight = FontWeight.Bold,
            color = fg,
        )
    }
}