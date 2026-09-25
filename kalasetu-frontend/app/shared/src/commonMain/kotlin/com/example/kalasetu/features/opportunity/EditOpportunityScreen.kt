package com.example.kalasetu.features.opportunity

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
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
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.launch

private val PurplePrimary = Color(0xFF7466F1)
private val TextDark @Composable get() = MaterialTheme.colorScheme.onSurface
private val TextGray @Composable get() = MaterialTheme.colorScheme.onSurfaceVariant
private val BorderGray @Composable get() = MaterialTheme.colorScheme.outlineVariant

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun EditOpportunityScreen(
    opportunityId: String,
    onBack: () -> Unit,
    onFinish: () -> Unit
) {
    val existingOpp = OpportunityStore.getOpportunity(opportunityId)
    val scope = rememberCoroutineScope()

    var title by remember { mutableStateOf(existingOpp?.title ?: "") }
    var location by remember { mutableStateOf(existingOpp?.location ?: "") }
    var description by remember { mutableStateOf(existingOpp?.description ?: "") }
    var totalPositions by remember { mutableStateOf(existingOpp?.totalPositions ?: 1) }
    var categories by remember { mutableStateOf(existingOpp?.categories ?: emptyList<String>()) }
    var isSubmitting by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text("ADMIN SUITE", fontSize = 10.sp, fontWeight = FontWeight.Bold, color = PurplePrimary)
                        Text("Edit Opportunity", fontSize = 20.sp, fontWeight = FontWeight.Bold)
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
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
                        onClick = onBack,
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, BorderGray)
                    ) {
                        Text("Cancel", color = TextDark)
                    }
                    Button(
                        onClick = {
                            if (!isSubmitting) {
                                isSubmitting = true
                                scope.launch {
                                    OpportunityRepository.updateOpportunity(
                                        id = opportunityId,
                                        title = title.ifBlank { "Updated Opportunity" },
                                        description = description,
                                        categories = categories,
                                        location = location,
                                        startDate = existingOpp?.startDate,
                                        endDate = existingOpp?.endDate,
                                        totalPositions = totalPositions,
                                        isDraft = false
                                    )
                                    isSubmitting = false
                                    onFinish()
                                }
                            }
                        },
                        modifier = Modifier.weight(1.2f),
                        shape = RoundedCornerShape(8.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = PurplePrimary),
                        enabled = !isSubmitting
                    ) {
                        if (isSubmitting) {
                            CircularProgressIndicator(modifier = Modifier.size(18.dp), color = Color.White, strokeWidth = 2.dp)
                        } else {
                            Text("Save Changes")
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

            SectionCard(index = "01", title = "Basic Information", isRequired = true) {
                Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                    CustomTextField(
                        label = "Opportunity Title *",
                        value = title,
                        onValueChange = { title = it },
                        placeholder = "e.g. Lead Guitarist Needed",
                        maxChar = 60
                    )

                    Column {
                        Text("Category & Focus *", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
                        Spacer(Modifier.height(8.dp))
                        FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            listOf("Music & Arts", "Community", "Environment", "Festival", "Charity").forEach { cat ->
                                CategoryChip(
                                    label = cat,
                                    isSelected = categories.contains(cat),
                                    onClick = {
                                        categories = if (categories.contains(cat)) categories - cat else categories + cat
                                    }
                                )
                            }
                        }
                    }

                    CustomTextField(
                        label = "Location or Venue *",
                        value = location,
                        onValueChange = { location = it },
                        placeholder = "e.g. Amphitheater Stage, Miami, FL",
                        leadingIcon = { Icon(Icons.Default.LocationOn, contentDescription = null, tint = TextGray) }
                    )
                }
            }

            SectionCard(index = "02", title = "Schedule & Capacity") {
                Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
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
                                Text("Positions open for applicants", fontSize = 11.sp, color = TextGray)
                            }
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                modifier = Modifier
                                    .clip(RoundedCornerShape(8.dp))
                                    .background(MaterialTheme.colorScheme.surface)
                                    .border(1.dp, BorderGray, RoundedCornerShape(8.dp))
                                    .padding(horizontal = 4.dp)
                            ) {
                                IconButton(onClick = { totalPositions = (totalPositions - 1).coerceAtLeast(1) }, modifier = Modifier.size(32.dp)) {
                                    Icon(Icons.Default.Remove, contentDescription = null, modifier = Modifier.size(18.dp))
                                }
                                Text(totalPositions.toString(), modifier = Modifier.padding(horizontal = 12.dp), fontWeight = FontWeight.Bold)
                                IconButton(onClick = { totalPositions += 1 }, modifier = Modifier.size(32.dp)) {
                                    Surface(color = PurplePrimary, shape = RoundedCornerShape(4.dp)) {
                                        Icon(Icons.Default.Add, contentDescription = null, tint = Color.White, modifier = Modifier.size(14.dp).padding(2.dp))
                                    }
                                }
                            }
                        }
                    }
                }
            }

            SectionCard(index = "03", title = "Description") {
                CustomTextField(
                    label = "Role Description *",
                    value = description,
                    onValueChange = { description = it },
                    placeholder = "Describe expectations, timing, and compensation...",
                    singleLine = false,
                    minHeight = 100.dp
                )
            }

            Spacer(Modifier.height(32.dp))
        }
    }
}
