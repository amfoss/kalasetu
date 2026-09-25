package com.example.kalasetu.features.opportunity

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel

private val PurplePrimary = Color(0xFF7466F1)
private val LightPurpleBg @Composable get() = MaterialTheme.colorScheme.primaryContainer
private val TextDark @Composable get() = MaterialTheme.colorScheme.onSurface
private val TextGray @Composable get() = MaterialTheme.colorScheme.onSurfaceVariant
private val BorderGray @Composable get() = MaterialTheme.colorScheme.outlineVariant

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun CreateOpportunityScreen(
    eventId: String,
    onBack: () -> Unit,
    onFinish: () -> Unit,
    viewModel: CreateOpportunityViewModel = viewModel()
) {
    val state by viewModel.state.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text("ADMIN SUITE", fontSize = 10.sp, fontWeight = FontWeight.Bold, color = PurplePrimary)
                        Text("Create Opportunity", fontSize = 20.sp, fontWeight = FontWeight.Bold)
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    Surface(
                        color = MaterialTheme.colorScheme.surfaceVariant,
                        shape = RoundedCornerShape(16.dp),
                        modifier = Modifier.padding(end = 12.dp)
                    ) {
                        Row(modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp), verticalAlignment = Alignment.CenterVertically) {
                            Icon(Icons.Default.Description, contentDescription = null, modifier = Modifier.size(14.dp), tint = TextGray)
                            Spacer(Modifier.width(4.dp))
                            Text("Draft", fontSize = 12.sp, color = TextGray)
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.surface)
            )
        },
        bottomBar = {
            Surface(shadowElevation = 8.dp, color = MaterialTheme.colorScheme.surface) {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(16.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    OutlinedButton(
                        onClick = { 
                            if (!state.isSubmitting) {
                                viewModel.submit(eventId, true) {
                                    onFinish()
                                }
                            }
                        },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, BorderGray)
                    ) {
                        if (state.isSubmitting) {
                            CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp)
                        } else {
                            Icon(Icons.Default.Save, contentDescription = null, modifier = Modifier.size(18.dp))
                            Spacer(Modifier.width(8.dp))
                            Text("Save Draft", color = TextDark)
                        }
                    }
                    Button(
                        onClick = { 
                            if (!state.isSubmitting) {
                                viewModel.submit(eventId, false) {
                                    onFinish()
                                }
                            }
                        },
                        modifier = Modifier.weight(1.2f),
                        shape = RoundedCornerShape(8.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = PurplePrimary)
                    ) {
                        if (state.isSubmitting) {
                            CircularProgressIndicator(modifier = Modifier.size(18.dp), color = Color.White, strokeWidth = 2.dp)
                        } else {
                            Text("Review & Publish")
                            Spacer(Modifier.width(8.dp))
                            Icon(Icons.AutoMirrored.Filled.ArrowForward, contentDescription = null, modifier = Modifier.size(18.dp))
                        }
                    }
                }
            }
        },
        containerColor = MaterialTheme.colorScheme.surfaceVariant
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Spacer(Modifier.height(8.dp))

            // Section 1: Basic Information
            SectionCard(index = "01", title = "Basic Information", isRequired = true) {
                Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                    CustomTextField(
                        label = "Opportunity Title *",
                        value = state.title,
                        onValueChange = viewModel::onTitleChange,
                        placeholder = "e.g. Sunset Music Festival Volunteer Crew",
                        maxChar = 60
                    )

                    Column {
                        Text("Category & Focus *", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
                        Spacer(Modifier.height(8.dp))
                        FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            listOf("Music & Arts", "Community", "Environment", "Festival", "Charity").forEach { cat ->
                                CategoryChip(
                                    label = cat,
                                    isSelected = state.categories.contains(cat),
                                    onClick = { viewModel.toggleCategory(cat) }
                                )
                            }
                        }
                    }

                    CustomTextField(
                        label = "Location or Venue *",
                        value = state.location,
                        onValueChange = viewModel::onLocationChange,
                        placeholder = "Bayfront Amphitheater, Miami, FL",
                        leadingIcon = { Icon(Icons.Default.LocationOn, contentDescription = null, tint = TextGray) }
                    )
                }
            }

            // Section 2: Schedule & Capacity
            SectionCard(index = "02", title = "Schedule & Capacity") {
                Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                        Column(modifier = Modifier.weight(1f)) {
                            Text("Start Date *", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
                            Spacer(Modifier.height(8.dp))
                            DateSelector(label = "Oct 24, 2025") { /* Pick date */ }
                        }
                        Column(modifier = Modifier.weight(1f)) {
                            Text("End Date *", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
                            Spacer(Modifier.height(8.dp))
                            DateSelector(label = "Oct 26, 2025") { /* Pick date */ }
                        }
                    }

                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(8.dp))
                            .background(MaterialTheme.colorScheme.surfaceVariant)
                            .padding(12.dp)
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Column {
                                Text("Total Available Positions", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
                                Text("Recommended 15-20 crew members", fontSize = 11.sp, color = TextGray)
                            }
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                modifier = Modifier
                                    .clip(RoundedCornerShape(8.dp))
                                    .background(MaterialTheme.colorScheme.surface)
                                    .border(1.dp, BorderGray, RoundedCornerShape(8.dp))
                                    .padding(horizontal = 4.dp)
                            ) {
                                IconButton(onClick = { viewModel.onPositionsChange(state.totalPositions - 1) }, modifier = Modifier.size(32.dp)) {
                                    Icon(Icons.Default.Remove, contentDescription = null, modifier = Modifier.size(18.dp))
                                }
                                Text(state.totalPositions.toString(), modifier = Modifier.padding(horizontal = 12.dp), fontWeight = FontWeight.Bold)
                                IconButton(onClick = { viewModel.onPositionsChange(state.totalPositions + 1) }, modifier = Modifier.size(32.dp)) {
                                    Surface(color = PurplePrimary, shape = RoundedCornerShape(4.dp)) {
                                        Icon(Icons.Default.Add, contentDescription = null, tint = Color.White, modifier = Modifier.size(14.dp).padding(2.dp))
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // Section 3: Description
            SectionCard(index = "03", title = "Description") {
                CustomTextField(
                    label = "Role Description *",
                    value = state.description,
                    onValueChange = viewModel::onDescriptionChange,
                    placeholder = "Outline the main responsibilities, team expectations, and volunteer vibe...",
                    singleLine = false,
                    minHeight = 100.dp
                )
            }

            Spacer(Modifier.height(32.dp))
        }
    }
}

@Composable
fun SectionCard(
    index: String,
    title: String,
    isRequired: Boolean = false,
    action: @Composable (() -> Unit)? = null,
    content: @Composable () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(24.dp)
                            .clip(CircleShape)
                            .background(LightPurpleBg),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(index, fontSize = 12.sp, fontWeight = FontWeight.Bold, color = PurplePrimary)
                    }
                    Spacer(Modifier.width(8.dp))
                    Text(title, fontSize = 18.sp, fontWeight = FontWeight.Bold, color = TextDark)
                    if (isRequired) {
                        Spacer(Modifier.width(8.dp))
                        Surface(color = Color(0xFFE0F2F1), shape = RoundedCornerShape(4.dp)) {
                            Text("REQUIRED", fontSize = 10.sp, fontWeight = FontWeight.Bold, color = Color(0xFF00796B), modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
                        }
                    }
                }
                action?.invoke()
            }
            Spacer(Modifier.height(16.dp))
            content()
        }
    }
}

@Composable
fun CategoryChip(label: String, isSelected: Boolean, onClick: () -> Unit) {
    Surface(
        onClick = onClick,
        color = if (isSelected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant,
        shape = RoundedCornerShape(20.dp),
        border = if (isSelected) null else BorderStroke(1.dp, BorderGray)
    ) {
        Row(modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp), verticalAlignment = Alignment.CenterVertically) {
            if (isSelected) {
                Icon(Icons.Default.MusicNote, contentDescription = null, tint = Color.White, modifier = Modifier.size(14.dp))
                Spacer(Modifier.width(4.dp))
            }
            Text(label, fontSize = 12.sp, fontWeight = FontWeight.Medium, color = if (isSelected) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
fun CustomTextField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    maxChar: Int? = null,
    leadingIcon: @Composable (() -> Unit)? = null,
    singleLine: Boolean = true,
    minHeight: Dp = 56.dp
) {
    Column {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Text(label, fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
            if (maxChar != null) {
                Text("${value.length}/$maxChar", fontSize = 12.sp, color = TextGray)
            }
        }
        Spacer(Modifier.height(8.dp))
        OutlinedTextField(
            value = value,
            onValueChange = onValueChange,
            placeholder = { Text(placeholder, fontSize = 14.sp, color = TextGray) },
            modifier = Modifier.fillMaxWidth().defaultMinSize(minHeight = minHeight),
            shape = RoundedCornerShape(8.dp),
            colors = OutlinedTextFieldDefaults.colors(
                unfocusedBorderColor = BorderGray,
                focusedBorderColor = PurplePrimary,
                unfocusedContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                focusedContainerColor = MaterialTheme.colorScheme.surfaceVariant
            ),
            leadingIcon = leadingIcon,
            singleLine = singleLine
        )
    }
}

@Composable
fun DateSelector(label: String, onClick: () -> Unit) {
    Surface(
        onClick = onClick,
        color = MaterialTheme.colorScheme.surfaceVariant,
        shape = RoundedCornerShape(8.dp),
        border = BorderStroke(1.dp, BorderGray)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 12.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(Icons.Default.CalendarToday, contentDescription = null, tint = TextGray, modifier = Modifier.size(18.dp))
            Spacer(Modifier.width(8.dp))
            Text(label, fontSize = 14.sp, color = TextDark)
        }
    }
}

