package com.example.kalasetu.features.application

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.CloudUpload
import androidx.compose.material.icons.filled.PictureAsPdf
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.core.utils.dashedBorder
import com.example.kalasetu.features.event.isValidEmail
import com.example.kalasetu.features.event.isValidPhone
import io.github.vinceglb.filekit.compose.rememberFilePickerLauncher
import io.github.vinceglb.filekit.core.PickerMode
import io.github.vinceglb.filekit.core.PickerType
import kotlinx.coroutines.launch
import kotlin.random.Random

private val PurplePrimary = Color(0xFF7466F1)
private val LightPurpleBg = Color(0xFFF4F1FF)
private val TextDark = Color(0xFF1E1E1E)
private val TextGray = Color(0xFF757575)
private val BorderGray = Color(0xFFE0E0E0)
private val SuccessGreen = Color(0xFF2E7D32)
private val SuccessGreenBg = Color(0xFFE5F7E5)
private val ErrorRed = Color(0xFFD32F2F)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ApplicationFormScreen(
    eventId: String,
    eventTitle: String,
    opportunityId: String = "",
    opportunityTitle: String = "",
    eventCoverBytes: ByteArray?,
    applicantAvatarBytes: ByteArray?,
    onBack: () -> Unit,
    onSubmit: (Application) -> Unit,
) {
    val scope = rememberCoroutineScope()

    var applicantName by remember { mutableStateOf("") }
    var reason by remember { mutableStateOf("") }
    var email by remember { mutableStateOf("") }
    var phone by remember { mutableStateOf("") }
    var portfolioBytes by remember { mutableStateOf<ByteArray?>(null) }
    var portfolioFileName by remember { mutableStateOf<String?>(null) }

    var showErrors by remember { mutableStateOf(false) }

    // ─── Validity checks ───
    val isNameValid = applicantName.trim().length >= 3
    val isReasonValid = reason.trim().length >= 10
    val isEmailValid = isValidEmail(email)
    val isPhoneValid = isValidPhone(phone)
    val isPortfolioValid = portfolioBytes != null

    val isFormValid = isNameValid && isReasonValid && isEmailValid && isPhoneValid && isPortfolioValid

    val portfolioFilePicker = rememberFilePickerLauncher(
        type = PickerType.File(extensions = listOf("pdf")),
        mode = PickerMode.Single,
    ) { file ->
        file?.let { picked ->
            scope.launch {
                try {
                    portfolioBytes = picked.readBytes()
                    portfolioFileName = picked.name
                } catch (e: Exception) {
                    println("Read failed: ${e.message}")
                }
            }
        }
    }

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(
                onClick = {
                    if (isFormValid) {
                        scope.launch {
                            val res = ApplicationRepository.submitApplication(
                                eventId = eventId,
                                opportunityId = opportunityId.ifBlank { null },
                                applicantName = applicantName.trim(),
                                email = email.trim(),
                                phone = phone.trim(),
                                description = reason.trim(),
                                resumeUrl = portfolioFileName ?: "portfolio.pdf"
                            )
                            val application = res.getOrNull() ?: Application(
                                id = "app_${Random.nextLong()}",
                                eventId = eventId,
                                eventTitle = eventTitle,
                                eventCoverBytes = eventCoverBytes,
                                applicantName = applicantName.trim(),
                                description = reason.trim(),
                                coverImageBytes = portfolioBytes,
                                applicantAvatarBytes = applicantAvatarBytes,
                                portfolioFileName = portfolioFileName ?: "",
                                email = email.trim(),
                                phone = phone.trim(),
                                status = ApplicationStatus.PENDING,
                            )
                            onSubmit(application)
                        }
                    } else {
                        showErrors = true
                    }
                },
                containerColor = if (isFormValid) LightPurpleBg else Color(0xFFEDEDED),
                contentColor = if (isFormValid) TextDark else Color(0xFF9E9E9E),
                shape = CircleShape,
                modifier = Modifier.size(56.dp),
            ) {
                Icon(
                    Icons.AutoMirrored.Filled.ArrowForward,
                    contentDescription = "Next",
                    modifier = Modifier.size(24.dp),
                )
            }
        },
        containerColor = Color.White,
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(horizontal = 24.dp)
                .verticalScroll(rememberScrollState()),
        ) {
            Spacer(Modifier.height(16.dp))

            IconButton(onClick = onBack, modifier = Modifier.offset(x = (-12).dp)) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back", tint = PurplePrimary)
            }

            Spacer(Modifier.height(8.dp))

            Text(
                "Let's start the\nNew Application",
                fontSize = 32.sp, fontWeight = FontWeight.Bold, color = TextDark, lineHeight = 38.sp,
            )
            Spacer(Modifier.height(12.dp))
            Text(
                "Add details to set up\nyour application.",
                fontSize = 24.sp, fontWeight = FontWeight.Normal, color = TextGray, lineHeight = 32.sp,
            )

            Spacer(Modifier.height(32.dp))

            // ─── Name ───
            AppValidatedField(
                label = "Name of the Applicant",
                placeholder = "e.g. Amrita Sharma",
                value = applicantName,
                onValueChange = { applicantName = it },
                errorMessage = when {
                    !showErrors && applicantName.isBlank() -> null
                    isNameValid -> null
                    else -> "Name must be at least 3 characters"
                },
            )
            Spacer(Modifier.height(16.dp))

            // ─── Reason ───
            AppValidatedField(
                label = "Why you want to apply for this event",
                placeholder = "Tell the organizer why you're a good fit...",
                value = reason,
                onValueChange = { reason = it },
                singleLine = false,
                minHeight = 120.dp,
                errorMessage = when {
                    !showErrors && reason.isBlank() -> null
                    isReasonValid -> null
                    else -> "Please write at least 10 characters"
                },
            )
            Spacer(Modifier.height(16.dp))

            // ─── Portfolio ───
            Text(
                "Upload Portfolio",
                fontSize = 14.sp, fontWeight = FontWeight.Medium, color = TextDark,
                modifier = Modifier.padding(bottom = 8.dp),
            )
            PortfolioUploadBox(
                previewBytes = portfolioBytes,
                fileName = portfolioFileName,
                onClick = { portfolioFilePicker.launch() },
            )
            if (showErrors && !isPortfolioValid) {
                Spacer(Modifier.height(6.dp))
                Text("Portfolio PDF is required", fontSize = 12.sp, color = ErrorRed)
            }
            Spacer(Modifier.height(16.dp))

            // ─── Email ───
            AppValidatedField(
                label = "Email",
                placeholder = "e.g. 123@345",
                value = email,
                onValueChange = { email = it },
                keyboardType = KeyboardType.Email,
                errorMessage = when {
                    !showErrors && email.isBlank() -> null
                    isEmailValid -> null
                    else -> "Enter a valid email (e.g. name@domain.com)"
                },
            )
            Spacer(Modifier.height(16.dp))

            // ─── Phone ───
            AppValidatedField(
                label = "Contact Number",
                placeholder = "e.g. 1234567890",
                value = phone,
                onValueChange = { new -> phone = new.filter { it.isDigit() }.take(10) },
                keyboardType = KeyboardType.Phone,
                errorMessage = when {
                    !showErrors && phone.isBlank() -> null
                    isPhoneValid -> null
                    else -> "Enter a valid 10-digit phone number"
                },
            )

            Spacer(Modifier.height(100.dp))
        }
    }
}

@Composable
private fun AppValidatedField(
    label: String,
    placeholder: String,
    value: String,
    onValueChange: (String) -> Unit,
    singleLine: Boolean = true,
    minHeight: Dp = 56.dp,
    keyboardType: KeyboardType = KeyboardType.Text,
    errorMessage: String? = null,
) {
    Column(Modifier.fillMaxWidth()) {
        Text(
            label,
            fontSize = 14.sp,
            fontWeight = FontWeight.Medium,
            color = TextDark,
            modifier = Modifier.padding(bottom = 8.dp),
        )
        OutlinedTextField(
            value = value,
            onValueChange = onValueChange,
            placeholder = { Text(placeholder, color = Color(0xFFBDBDBD), fontSize = 14.sp) },
            modifier = Modifier
                .fillMaxWidth()
                .defaultMinSize(minHeight = minHeight),
            shape = RoundedCornerShape(8.dp),
            singleLine = singleLine,
            colors = OutlinedTextFieldDefaults.colors(
                unfocusedBorderColor = if (errorMessage != null) ErrorRed else BorderGray,
                focusedBorderColor = if (errorMessage != null) ErrorRed else PurplePrimary,
                unfocusedContainerColor = Color.White,
                focusedContainerColor = Color.White,
            ),
            isError = errorMessage != null,
            keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
        )
        if (errorMessage != null) {
            Spacer(Modifier.height(4.dp))
            Text(errorMessage, fontSize = 12.sp, color = ErrorRed)
        }
    }
}

@Composable
private fun PortfolioUploadBox(
    previewBytes: ByteArray?,
    fileName: String?,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(LightPurpleBg)
            .dashedBorder(color = PurplePrimary, strokeWidth = 1.5.dp, cornerRadius = 12.dp)
            .clickable { onClick() }
            .padding(if (previewBytes != null) 12.dp else 24.dp),
        contentAlignment = Alignment.Center,
    ) {
        if (previewBytes == null) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Box(
                    modifier = Modifier.size(48.dp).clip(CircleShape).background(PurplePrimary),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Default.CloudUpload, "Upload", tint = Color.White, modifier = Modifier.size(24.dp))
                }
                Spacer(Modifier.height(12.dp))
                Text("Upload Your Portfolio", fontSize = 16.sp, fontWeight = FontWeight.Bold, color = TextDark)
                Spacer(Modifier.height(4.dp))
                Text(
                    "Supported Format: PDF only\n(max 10 MB)",
                    fontSize = 12.sp, color = TextGray, textAlign = TextAlign.Center, lineHeight = 16.sp,
                )
            }
        } else {
            Row(
                modifier = Modifier.fillMaxWidth().padding(4.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Box(
                    modifier = Modifier.size(48.dp).clip(RoundedCornerShape(8.dp)).background(Color.White),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Default.PictureAsPdf, null, tint = PurplePrimary, modifier = Modifier.size(26.dp))
                }
                Spacer(Modifier.width(12.dp))
                Column(Modifier.weight(1f)) {
                    Text(
                        fileName ?: "Portfolio.pdf",
                        fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TextDark, maxLines = 1,
                    )
                    Spacer(Modifier.height(2.dp))
                    Text("PDF • Uploaded", fontSize = 12.sp, color = TextGray)
                }
                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(50))
                        .background(SuccessGreenBg)
                        .padding(horizontal = 8.dp, vertical = 4.dp),
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.Check, null, tint = SuccessGreen, modifier = Modifier.size(12.dp))
                        Spacer(Modifier.width(4.dp))
                        Text("Uploaded", fontSize = 10.sp, color = SuccessGreen, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}
