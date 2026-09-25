package com.example.kalasetu.features.application

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Email
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Phone
import androidx.compose.material.icons.filled.PictureAsPdf
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
import io.github.vinceglb.filekit.compose.rememberFileSaverLauncher

private val PurplePrimary = Color(0xFF7466F1)
private val LightPurpleBg @Composable get() = MaterialTheme.colorScheme.primaryContainer
private val TextDark @Composable get() = MaterialTheme.colorScheme.onSurface
private val TextGray @Composable get() = MaterialTheme.colorScheme.onSurfaceVariant
private val SuccessGreen = Color(0xFF2E7D32)
private val SuccessGreenBg = Color(0xFFE5F7E5)
private val RejectRed = Color(0xFFD32F2F)
private val RejectRedBg = Color(0xFFFFE5E5)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ApplicationPreviewScreen(
    application: Application,
    onBack: () -> Unit,
    onAccept: () -> Unit,
    onReject: () -> Unit,
    onDone: () -> Unit,
) {
    var saveMessage by remember { mutableStateOf<String?>(null) }

    // FileKit save launcher — opens the OS "save file" dialog
    val fileSaver = rememberFileSaverLauncher { savedFile ->
        saveMessage = if (savedFile != null) "Saved successfully" else "Save cancelled"
    }

    // Read live status from store so button action reflects immediately
    val allApps by ApplicationStore.applications.collectAsState()
    val liveApp = allApps.firstOrNull { it.id == application.id } ?: application

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Application Preview", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
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
                .padding(horizontal = 20.dp),
        ) {
            Spacer(Modifier.height(8.dp))

            // ─── Basic Information Card ───
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = LightPurpleBg),
                elevation = CardDefaults.cardElevation(0.dp),
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        // Profile avatar from the applicant's profile
                        Box(
                            modifier = Modifier
                                .size(84.dp)
                                .clip(CircleShape)
                                .background(MaterialTheme.colorScheme.surface),
                            contentAlignment = Alignment.Center,
                        ) {
                            if (liveApp.applicantAvatarBytes != null) {
                                AsyncImage(
                                    model = liveApp.applicantAvatarBytes,
                                    contentDescription = "Applicant avatar",
                                    contentScale = ContentScale.Crop,
                                    modifier = Modifier.fillMaxSize().clip(CircleShape),
                                )
                            } else {
                                Text(
                                    liveApp.applicantName.take(1).uppercase().ifBlank { "A" },
                                    fontSize = 32.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = PurplePrimary,
                                )
                            }
                        }

                        Spacer(Modifier.width(16.dp))

                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier
                                    .size(28.dp)
                                    .clip(CircleShape)
                                    .background(PurplePrimary.copy(alpha = 0.15f)),
                                contentAlignment = Alignment.Center,
                            ) {
                                Icon(
                                    Icons.Default.Person,
                                    null,
                                    tint = PurplePrimary,
                                    modifier = Modifier.size(16.dp),
                                )
                            }
                            Spacer(Modifier.width(8.dp))
                            Text(
                                "Basic Information",
                                fontSize = 16.sp,
                                fontWeight = FontWeight.Bold,
                                color = TextDark,
                            )
                        }
                    }

                    Spacer(Modifier.height(16.dp))

                    InfoRow(Icons.Default.Person, "Name", liveApp.applicantName.ifBlank { "—" })
                    InfoRow(Icons.Default.Email, "Email", liveApp.email.ifBlank { "—" })
                    InfoRow(Icons.Default.Phone, "Contact", liveApp.phone.ifBlank { "—" })
                }
            }

            Spacer(Modifier.height(20.dp))

            // ─── Reason ───
            if (liveApp.description.isNotBlank()) {
                Text("Why they want to apply", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark)
                Spacer(Modifier.height(8.dp))
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(12.dp))
                        .background(LightPurpleBg)
                        .padding(14.dp),
                ) {
                    Text(liveApp.description, fontSize = 13.sp, color = TextDark, lineHeight = 18.sp)
                }
                Spacer(Modifier.height(20.dp))
            }

            // ─── Attachments ───
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text("Attachments", fontSize = 16.sp, fontWeight = FontWeight.Bold, color = TextDark)
                if (liveApp.coverImageBytes != null) {
                    Box(
                        modifier = Modifier
                            .clip(RoundedCornerShape(50))
                            .background(PurplePrimary)
                            .padding(horizontal = 12.dp, vertical = 4.dp),
                    ) {
                        Text("1 file", fontSize = 11.sp, color = Color.White, fontWeight = FontWeight.Bold)
                    }
                }
            }

            Spacer(Modifier.height(12.dp))

            if (liveApp.coverImageBytes != null) {
                val fileName = liveApp.portfolioFileName.ifBlank { "Portfolio" }
                val ext = fileName.substringAfterLast('.', "").lowercase()

                AttachmentRow(
                    fileName = fileName,
                    extension = ext.ifBlank { "file" },
                    onDownload = {
                        // ✅ Use the save launcher — opens OS save dialog
                        val bytes = liveApp.coverImageBytes
                        if (bytes != null) {
                            saveMessage = "Opening save dialog..."
                            fileSaver.launch(
                                bytes = bytes,
                                baseName = fileName.substringBeforeLast('.', fileName),
                                extension = ext.ifBlank { "pdf" },
                            )
                        }
                    },
                )
            } else {
                Text("No attachments", fontSize = 13.sp, color = TextGray)
            }

            if (saveMessage != null) {
                Spacer(Modifier.height(8.dp))
                Text(saveMessage!!, fontSize = 12.sp, color = PurplePrimary)
            }

            Spacer(Modifier.height(32.dp))

            // ─── Status actions ───
            when (liveApp.status) {
                ApplicationStatus.PENDING -> {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                        Button(
                            onClick = onReject,
                            modifier = Modifier.weight(1f).height(52.dp),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = RejectRedBg),
                        ) {
                            Icon(Icons.Default.Close, null, tint = RejectRed, modifier = Modifier.size(20.dp))
                            Spacer(Modifier.width(6.dp))
                            Text("Reject", color = RejectRed, fontWeight = FontWeight.Bold)
                        }
                        Button(
                            onClick = onAccept,
                            modifier = Modifier.weight(1f).height(52.dp),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = SuccessGreenBg),
                        ) {
                            Icon(Icons.Default.Check, null, tint = SuccessGreen, modifier = Modifier.size(20.dp))
                            Spacer(Modifier.width(6.dp))
                            Text("Accept", color = SuccessGreen, fontWeight = FontWeight.Bold)
                        }
                    }
                    Spacer(Modifier.height(12.dp))
                }
                ApplicationStatus.ACCEPTED -> {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(12.dp))
                            .background(SuccessGreenBg)
                            .padding(14.dp),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text("✅ This application is Accepted", color = SuccessGreen, fontWeight = FontWeight.Bold)
                    }
                    Spacer(Modifier.height(12.dp))
                }
                ApplicationStatus.REJECTED -> {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(12.dp))
                            .background(RejectRedBg)
                            .padding(14.dp),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text("❌ This application is Rejected", color = RejectRed, fontWeight = FontWeight.Bold)
                    }
                    Spacer(Modifier.height(12.dp))
                }
            }

            Button(
                onClick = onDone,
                modifier = Modifier.fillMaxWidth().height(52.dp),
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
private fun InfoRow(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    value: String,
) {
    Row(
        modifier = Modifier.fillMaxWidth().padding(vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(icon, null, tint = PurplePrimary, modifier = Modifier.size(18.dp))
        Spacer(Modifier.width(12.dp))
        Text(label, fontSize = 13.sp, color = TextGray, modifier = Modifier.width(88.dp))
        Text(value, fontSize = 13.sp, fontWeight = FontWeight.Medium, color = TextDark, modifier = Modifier.weight(1f))
    }
}

@Composable
private fun AttachmentRow(
    fileName: String,
    extension: String,
    onDownload: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(LightPurpleBg)
            .padding(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(44.dp)
                .clip(RoundedCornerShape(8.dp))
                .background(RejectRedBg),
            contentAlignment = Alignment.Center,
        ) {
            Icon(Icons.Default.PictureAsPdf, null, tint = RejectRed, modifier = Modifier.size(24.dp))
        }
        Spacer(Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(fileName, fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark, maxLines = 1)
            Spacer(Modifier.height(2.dp))
            Text(extension.uppercase(), fontSize = 12.sp, color = TextGray)
        }
        Box(
            modifier = Modifier
                .size(36.dp)
                .clip(RoundedCornerShape(8.dp))
                .background(MaterialTheme.colorScheme.surface)
                .clickable { onDownload() },
            contentAlignment = Alignment.Center,
        ) {
            Icon(Icons.Default.Download, "Download", tint = PurplePrimary, modifier = Modifier.size(18.dp))
        }
    }
}