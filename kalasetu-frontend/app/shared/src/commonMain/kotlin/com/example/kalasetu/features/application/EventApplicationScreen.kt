package com.example.kalasetu.features.application

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.KeyboardArrowRight
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil3.compose.AsyncImage
import com.example.kalasetu.core.utils.dashedBorder
import com.example.kalasetu.features.event.Event
import com.example.kalasetu.features.opportunity.Opportunity
import com.example.kalasetu.features.opportunity.OpportunityRepository
import com.example.kalasetu.features.opportunity.OpportunityStatus
import com.example.kalasetu.features.opportunity.OpportunityStore
import kotlin.time.Clock
import kotlinx.datetime.LocalDate
import kotlinx.datetime.daysUntil

private val PurplePrimary = Color(0xFF7466F1)
private val LightPurpleBg = Color(0xFFF4F1FF)
private val TextDark = Color(0xFF1E1E1E)
private val TextGray = Color(0xFF757575)
private val BorderGray = Color(0xFFE0E0E0)
private val SuccessGreen = Color(0xFF2E7D32)
private val SuccessGreenBg = Color(0xFFE5F7E5)

private val StatusPendingBg = Color(0xFFFFF4E5)
private val StatusPendingFg = Color(0xFFE69500)
private val StatusAcceptedBg = Color(0xFFE5F7E5)
private val StatusAcceptedFg = Color(0xFF2E7D32)
private val StatusRejectedBg = Color(0xFFFFE5E5)
private val StatusRejectedFg = Color(0xFFD32F2F)

@Composable
fun EventApplicationsScreen(
    event: Event,
    onBack: () -> Unit,
    onApplicationClick: (String) -> Unit,
    onCreateOpportunity: () -> Unit = {},
    onOpportunityClick: (String) -> Unit = {},
) {
    LaunchedEffect(event.id) {
        ApplicationRepository.fetchApplicationsForEvent(event.id)
        OpportunityRepository.fetchOpportunitiesForEvent(event.id)
    }

    // Live read from the shared store — updates the moment an artist applies
    val allApps by ApplicationStore.applications.collectAsState()
    val eventApps = remember(allApps, event.id) {
        allApps.filter { it.eventId == event.id }
    }

    val allOpps by OpportunityStore.opportunities.collectAsState()
    val eventOpps = remember(allOpps, event.id) {
        allOpps.filter { it.eventId == event.id }
    }

    var selectedMainTab by remember { mutableStateOf(0) } // 0: Applications, 1: Opportunity

    val tabs = listOf("All", "Pending", "Accepted", "Rejected")
    var selectedAppTab by remember { mutableStateOf(0) }

    val filteredApps = remember(eventApps, selectedAppTab) {
        when (selectedAppTab) {
            1 -> eventApps.filter { it.status == ApplicationStatus.PENDING }
            2 -> eventApps.filter { it.status == ApplicationStatus.ACCEPTED }
            3 -> eventApps.filter { it.status == ApplicationStatus.REJECTED }
            else -> eventApps
        }
    }

    val appCounts = listOf(
        eventApps.size,
        eventApps.count { it.status == ApplicationStatus.PENDING },
        eventApps.count { it.status == ApplicationStatus.ACCEPTED },
        eventApps.count { it.status == ApplicationStatus.REJECTED },
    )

    Scaffold(containerColor = Color.White) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
        ) {
            Spacer(Modifier.height(8.dp))

            // ─── Back arrow ───
            IconButton(
                onClick = onBack,
                modifier = Modifier.offset(x = (-12).dp),
            ) {
                Icon(
                    Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                    tint = Color.Black,
                )
            }

            // ─── Cover image with 3 info pills ───
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(220.dp)
                    .clip(RoundedCornerShape(16.dp)),
            ) {
                if (event.coverImageBytes != null) {
                    AsyncImage(
                        model = event.coverImageBytes,
                        contentDescription = event.title,
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
                        Icon(
                            Icons.Default.Image, null,
                            tint = PurplePrimary,
                            modifier = Modifier.size(48.dp),
                        )
                    }
                }

                // Overlay: 3 pills at bottom of cover
                Row(
                    modifier = Modifier
                        .align(Alignment.BottomStart)
                        .fillMaxWidth()
                        .padding(10.dp),
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    InfoPill(
                        icon = Icons.Default.CalendarToday,
                        label = "Duration",
                        value = formatDurationPill(event),
                        modifier = Modifier.weight(1f),
                    )
                    InfoPill(
                        icon = Icons.Default.Person,
                        label = "Organizer",
                        value = event.organizerName.ifBlank { "KalaSetu" },
                        modifier = Modifier.weight(1f),
                    )
                    InfoPill(
                        icon = Icons.Default.LocationOn,
                        label = "Location",
                        value = event.location.ifBlank { "—" },
                        modifier = Modifier.weight(1f),
                    )
                }
            }

            Spacer(Modifier.height(16.dp))

            // ─── Title + description ───
            Text(
                event.title.ifBlank { "Untitled Event" },
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark,
            )
            Spacer(Modifier.height(4.dp))
            Text(
                event.description.ifBlank { "No description provided" },
                fontSize = 13.sp,
                color = TextGray,
                lineHeight = 18.sp,
                maxLines = 2,
            )

            Spacer(Modifier.height(16.dp))

            // ─── Main Tabs (Applications vs Opportunity) ───
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(RoundedCornerShape(12.dp))
                    .background(Color(0xFFF5F5F5))
                    .padding(4.dp),
            ) {
                MainTabItem(
                    label = "Applications (${eventApps.size})",
                    isSelected = selectedMainTab == 0,
                    modifier = Modifier.weight(1f),
                    onClick = { selectedMainTab = 0 }
                )
                MainTabItem(
                    label = "Opportunity (${eventOpps.size})",
                    isSelected = selectedMainTab == 1,
                    modifier = Modifier.weight(1f),
                    onClick = { selectedMainTab = 1 }
                )
            }

            Spacer(Modifier.height(20.dp))

            if (selectedMainTab == 0) {
                // ─── Applications Content ───
                ApplicationsTabContent(
                    tabs = tabs,
                    selectedTab = selectedAppTab,
                    onTabSelected = { selectedAppTab = it },
                    appCounts = appCounts,
                    filteredApps = filteredApps,
                    eventApps = eventApps,
                    onApplicationClick = onApplicationClick
                )
            } else {
                // ─── Opportunity Content ───
                OpportunityTabContent(
                    opportunities = eventOpps,
                    onCreateOpportunity = onCreateOpportunity,
                    onOpportunityClick = onOpportunityClick
                )
            }
        }
    }
}

@Composable
private fun MainTabItem(
    label: String,
    isSelected: Boolean,
    modifier: Modifier = Modifier,
    onClick: () -> Unit
) {
    Box(
        modifier = modifier
            .clip(RoundedCornerShape(8.dp))
            .background(if (isSelected) PurplePrimary else Color.Transparent)
            .clickable { onClick() }
            .padding(vertical = 12.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = label,
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            color = if (isSelected) Color.White else TextGray
        )
    }
}

@Composable
private fun ApplicationsTabContent(
    tabs: List<String>,
    selectedTab: Int,
    onTabSelected: (Int) -> Unit,
    appCounts: List<Int>,
    filteredApps: List<Application>,
    eventApps: List<Application>,
    onApplicationClick: (String) -> Unit
) {
    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(
                "Applications",
                fontSize = 20.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark
            )
            Text(
                "Sort: Recent",
                fontSize = 12.sp,
                color = PurplePrimary,
                fontWeight = FontWeight.Medium
            )
        }
        
        Spacer(Modifier.height(12.dp))

        // ─── Sub-tabs (All, Pending, etc.) ───
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            tabs.forEachIndexed { i, label ->
                val isSelected = selectedTab == i
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .clip(RoundedCornerShape(20.dp))
                        .background(if (isSelected) PurplePrimary else Color(0xFFF5F5F5))
                        .clickable { onTabSelected(i) }
                        .padding(vertical = 8.dp),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = "$label (${appCounts[i]})",
                        fontSize = 10.sp,
                        fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                        color = if (isSelected) Color.White else TextGray,
                        maxLines = 1,
                    )
                }
            }
        }

        Spacer(Modifier.height(16.dp))

        if (filteredApps.isEmpty()) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        "0 Applications",
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                        color = TextDark,
                    )
                    Spacer(Modifier.height(6.dp))
                    Text(
                        if (eventApps.isEmpty()) "No artists have applied yet" else "No results in this tab",
                        fontSize = 13.sp,
                        color = TextGray,
                    )
                }
            }
        } else {
            LazyColumn(
                verticalArrangement = Arrangement.spacedBy(12.dp),
                contentPadding = PaddingValues(bottom = 24.dp),
            ) {
                items(filteredApps, key = { it.id }) { app ->
                    OrganizerAppRow(
                        app = app,
                        onClick = { onApplicationClick(app.id) },
                    )
                }
            }
        }
    }
}

@Composable
private fun OpportunityTabContent(
    opportunities: List<Opportunity>,
    onCreateOpportunity: () -> Unit,
    onOpportunityClick: (String) -> Unit
) {
    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column {
                Text(
                    "Active Opportunities",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark
                )
                Text(
                    "${opportunities.size} listings managed",
                    fontSize = 12.sp,
                    color = TextGray
                )
            }
            Button(
                onClick = onCreateOpportunity,
                colors = ButtonDefaults.buttonColors(containerColor = PurplePrimary),
                shape = RoundedCornerShape(8.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp)
            ) {
                Text("Create Opportunity", fontSize = 12.sp, fontWeight = FontWeight.Bold)
            }
        }

        Spacer(Modifier.height(20.dp))

        if (opportunities.isEmpty()) {
            Box(modifier = Modifier.fillMaxWidth().height(200.dp), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("No opportunities created", color = TextGray)
                    Spacer(Modifier.height(12.dp))
                    AddAnotherOpportunityButton(onClick = onCreateOpportunity)
                }
            }
        } else {
            LazyColumn(
                verticalArrangement = Arrangement.spacedBy(16.dp),
                contentPadding = PaddingValues(bottom = 24.dp)
            ) {
                items(opportunities, key = { it.id }) { opp ->
                    OpportunityCard(opp, onClick = { onOpportunityClick(opp.id) })
                }
                
                item {
                    AddAnotherOpportunityButton(onClick = onCreateOpportunity)
                }
            }
        }
    }
}

@Composable
private fun OpportunityCard(opp: Opportunity, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, BorderGray)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        opp.title,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                        color = TextDark
                    )
                    Spacer(Modifier.width(8.dp))
                    StatusBadge(opp.status)
                }
                Row(verticalAlignment = Alignment.CenterVertically) {
                    IconButton(onClick = onClick, modifier = Modifier.size(28.dp)) {
                        Icon(Icons.Default.Edit, contentDescription = "Edit Opportunity", tint = PurplePrimary, modifier = Modifier.size(20.dp))
                    }
                }
            }
            
            Spacer(Modifier.height(4.dp))
            Text(
                opp.description,
                fontSize = 13.sp,
                color = TextGray,
                maxLines = 2,
                lineHeight = 18.sp
            )
            
            Spacer(Modifier.height(16.dp))
            
            // Stats Row
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(RoundedCornerShape(8.dp))
                    .background(Color(0xFFF8F9FA))
                    .padding(vertical = 12.dp),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                StatItem(label = "Applications", value = opp.applicationsCount.toString())
                VerticalDivider(modifier = Modifier.height(24.dp), color = BorderGray)
                StatItem(label = "Open Slots", value = opp.openSlots.toString())
                VerticalDivider(modifier = Modifier.height(24.dp), color = BorderGray)
                StatItem(label = "Remaining", value = formatRemaining(opp.endDate))
            }
            
            Spacer(Modifier.height(12.dp))
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    "View Details",
                    fontSize = 12.sp,
                    color = TextDark,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier.clickable { /* View Details */ }
                )
                Spacer(Modifier.width(16.dp))
                if (opp.status == OpportunityStatus.DRAFT) {
                    Button(
                        onClick = { /* Publish */ },
                        colors = ButtonDefaults.buttonColors(containerColor = PurplePrimary),
                        shape = RoundedCornerShape(6.dp),
                        contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                        modifier = Modifier.height(32.dp)
                    ) {
                        Text("Publish", fontSize = 12.sp)
                    }
                }
                Spacer(Modifier.width(8.dp))
                OutlinedButton(
                    onClick = { /* Edit */ },
                    shape = RoundedCornerShape(6.dp),
                    border = BorderStroke(1.dp, BorderGray),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                    modifier = Modifier.height(32.dp)
                ) {
                    Text("Edit", fontSize = 12.sp, color = TextDark)
                }
            }
        }
    }
}

@Composable
private fun StatItem(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(value, fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
        Text(label, fontSize = 10.sp, color = TextGray)
    }
}

@Composable
private fun StatusBadge(status: OpportunityStatus) {
    val (bg, fg) = when (status) {
        OpportunityStatus.ACTIVE -> SuccessGreenBg to SuccessGreen
        OpportunityStatus.DRAFT -> Color(0xFFFFF8E1) to Color(0xFFFBC02D)
    }
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(4.dp))
            .background(bg)
            .padding(horizontal = 6.dp, vertical = 2.dp)
    ) {
        Text(status.label, fontSize = 10.sp, fontWeight = FontWeight.Bold, color = fg)
    }
}

@Composable
private fun AddAnotherOpportunityButton(onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .dashedBorder(color = PurplePrimary, strokeWidth = 1.dp, cornerRadius = 8.dp)
            .clickable { onClick() }
            .padding(vertical = 12.dp),
        contentAlignment = Alignment.Center
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Icon(Icons.Default.AddCircleOutline, contentDescription = null, tint = PurplePrimary, modifier = Modifier.size(20.dp))
            Spacer(Modifier.width(8.dp))
            Text("+ Add Another Opportunity", color = PurplePrimary, fontSize = 14.sp, fontWeight = FontWeight.Medium)
        }
    }
}

private fun formatRemaining(endDate: LocalDate?): String {
    if (endDate == null) return "—"
    val days = (endDate.toEpochDays() - (Clock.System.now().toEpochMilliseconds() / 86_400_000L)).toInt()
    return if (days > 0) "${days}d" else "Ends today"
}

@Composable
private fun InfoPill(
    icon: ImageVector,
    label: String,
    value: String,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .clip(RoundedCornerShape(8.dp))
            .background(Color.White)
            .padding(horizontal = 8.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(
            icon,
            contentDescription = null,
            tint = PurplePrimary,
            modifier = Modifier.size(14.dp),
        )
        Spacer(Modifier.width(6.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                label,
                fontSize = 9.sp,
                color = TextGray,
                maxLines = 1,
            )
            Text(
                value,
                fontSize = 10.sp,
                color = TextDark,
                fontWeight = FontWeight.Bold,
                maxLines = 1,
                softWrap = false,
            )
        }
    }
}

@Composable
private fun OrganizerAppRow(app: Application, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = LightPurpleBg),
        elevation = CardDefaults.cardElevation(0.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            // Avatar / portfolio photo
            Box(
                modifier = Modifier
                    .size(52.dp)
                    .clip(RoundedCornerShape(8.dp))
                    .background(Color.White),
                contentAlignment = Alignment.Center,
            ) {
                if (app.applicantAvatarBytes != null) {
                    AsyncImage(
                        model = app.applicantAvatarBytes,
                        contentDescription = null,
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize().clip(CircleShape),
                    )
                } else {
                    Text(
                        app.applicantName.take(1).uppercase().ifBlank { "A" },
                        fontSize = 22.sp,
                        fontWeight = FontWeight.Bold,
                        color = PurplePrimary,
                    )
                }
            }

            Spacer(Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    app.applicantName.ifBlank { "Applicant" },
                    fontSize = 15.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark,
                )
                Spacer(Modifier.height(2.dp))
                Text(
                    app.description.ifBlank { "No description" },
                    fontSize = 12.sp,
                    color = TextGray,
                    maxLines = 2,
                    lineHeight = 16.sp,
                )
            }

            Spacer(Modifier.width(8.dp))

            StatusBadge(status = app.status)

            Spacer(Modifier.width(6.dp))

            Icon(
                Icons.AutoMirrored.Filled.KeyboardArrowRight,
                contentDescription = null,
                tint = TextDark,
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

private fun formatDurationPill(event: Event): String {
    val start = event.startDate ?: return "TBA"
    val end = event.endDate
    if (end == null || end == start) return "1 day"
    val days = start.daysUntil(end) + 1
    return "$days days"
}
