package com.example.kalasetu.features.event

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
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.CloudUpload
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.PhotoLibrary
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.material.icons.filled.Menu
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import coil3.compose.AsyncImage
import com.example.kalasetu.core.utils.dashedBorder
import io.github.vinceglb.filekit.compose.rememberFilePickerLauncher
import io.github.vinceglb.filekit.core.PickerMode
import io.github.vinceglb.filekit.core.PickerType
import kotlinx.coroutines.launch


// ─── Shared validators (used by ApplicationFormScreen too) ───
internal fun isValidEmail(email: String): Boolean =
    email.trim().matches(Regex("^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$"))

internal fun isValidPhone(phone: String): Boolean =
    phone.filter { it.isDigit() }.length == 10

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun CreateEventScreen(
    initialDraft: EventDraft,
    viewModel: CreateEventViewModel = viewModel(),
    onNext: (
        title: String,
        description: String,
        email: String,
        phone: String,
        coverImageBytes: ByteArray?,
        galleryBytes: List<ByteArray>
    ) -> Unit,
    onBack: () -> Unit,
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()

    var coverBytes by remember { mutableStateOf<ByteArray?>(initialDraft.coverImageBytes) }
    var galleryBytes by remember { mutableStateOf<List<ByteArray>>(initialDraft.galleryBytes) }
    var showSourcePicker by remember { mutableStateOf(false) }
    var pickerTarget by remember { mutableStateOf("COVER") }

    // ─── Validation state ───
    var showErrors by remember { mutableStateOf(false) }

    val isTitleValid = state.eventName.trim().length >= 3
    val isDescriptionValid = state.description.trim().length >= 10
    val isEmailValid = isValidEmail(state.email)
    val isPhoneValid = isValidPhone(state.contactNumber)
    val isCoverValid = coverBytes != null

    val isFormValid = isTitleValid && isDescriptionValid && isEmailValid && isPhoneValid && isCoverValid

    LaunchedEffect(initialDraft) {
        viewModel.syncFromDraft(
            title = initialDraft.title,
            description = initialDraft.description,
            email = initialDraft.email,
            phone = initialDraft.phone,
            coverImageName = null,
        )
    }

    // ─── Cover pickers ───
    val coverImagePicker = rememberFilePickerLauncher(
        type = PickerType.Image,
        mode = PickerMode.Single,
    ) { file ->
        file?.let { picked ->
            viewModel.onCoverImageChange(picked.name)
            scope.launch {
                try { coverBytes = picked.readBytes() }
                catch (e: Exception) { println("Read failed: ${e.message}") }
            }
        }
    }
    val coverFilePicker = rememberFilePickerLauncher(
        type = PickerType.File(extensions = listOf("pdf", "png", "jpg", "jpeg", "doc", "docx")),
        mode = PickerMode.Single,
    ) { file ->
        file?.let { picked ->
            viewModel.onCoverImageChange(picked.name)
            scope.launch {
                try { coverBytes = picked.readBytes() }
                catch (e: Exception) { println("Read failed: ${e.message}") }
            }
        }
    }

    // ─── Gallery pickers ───
    val galleryImagePicker = rememberFilePickerLauncher(
        type = PickerType.Image,
        mode = PickerMode.Multiple(),
    ) { files ->
        if (!files.isNullOrEmpty()) {
            viewModel.onGalleryImagesChange(files.map { it.name })
            scope.launch {
                val bytes = files.mapNotNull { try { it.readBytes() } catch (e: Exception) { null } }
                galleryBytes = galleryBytes + bytes
            }
        }
    }
    val galleryFilePicker = rememberFilePickerLauncher(
        type = PickerType.File(extensions = listOf("png", "jpg", "jpeg", "webp")),
        mode = PickerMode.Multiple(),
    ) { files ->
        if (!files.isNullOrEmpty()) {
            viewModel.onGalleryImagesChange(files.map { it.name })
            scope.launch {
                val bytes = files.mapNotNull { try { it.readBytes() } catch (e: Exception) { null } }
                galleryBytes = galleryBytes + bytes
            }
        }
    }

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(
                onClick = {
                    if (isFormValid) {
                        viewModel.submitEvent()
                        onNext(
                            state.eventName.trim(),
                            state.description.trim(),
                            state.email.trim(),
                            state.contactNumber.trim(),
                            coverBytes,
                            galleryBytes,
                        )
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
                    imageVector = Icons.AutoMirrored.Filled.ArrowForward,
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
                .verticalScroll(rememberScrollState())
        ) {
            Spacer(modifier = Modifier.height(16.dp))

            IconButton(
                onClick = onBack,
                modifier = Modifier.padding(bottom = 8.dp)
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                    tint = PurplePrimary,
                )
            }

            Text(
                text = "Let's start\nCreating a new\nEvent",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark,
                lineHeight = 38.sp,
            )

            Spacer(modifier = Modifier.height(32.dp))

            // ─── Event Name ───
            CustomTextField(
                label = "What's Name of the Event",
                placeholder = "e.g. Summer Struck",
                value = state.eventName,
                onValueChange = viewModel::onEventNameChange,
                errorMessage = when {
                    !showErrors && state.eventName.isBlank() -> null
                    isTitleValid -> null
                    else -> "Name must be at least 3 characters"
                },
            )

            Spacer(modifier = Modifier.height(16.dp))

            // ─── Description ───
            CustomTextField(
                label = "Description",
                placeholder = "Enter the Description",
                value = state.description,
                onValueChange = viewModel::onDescriptionChange,
                singleLine = false,
                minHeight = 140.dp,
                errorMessage = when {
                    !showErrors && state.description.isBlank() -> null
                    isDescriptionValid -> null
                    else -> "Description must be at least 10 characters"
                },
            )

            Spacer(modifier = Modifier.height(16.dp))

            // ─── Cover ───
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Bottom,
            ) {
                Text(
                    text = "Upload cover page",
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Medium,
                    color = TextDark,
                    modifier = Modifier.padding(bottom = 8.dp),
                )
                Text(
                    text = "16:9 recommended",
                    fontSize = 11.sp,
                    color = PurplePrimary,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier.padding(bottom = 10.dp),
                )
            }
            UploadBox(
                title = state.coverImageName ?: "Upload Your Cover Page",
                subtitle = "Supported Format: PDF, JPG, PNG (Max 10MB)",
                previewBytes = coverBytes,
            ) {
                pickerTarget = "COVER"
                showSourcePicker = true
            }
            if (showErrors && !isCoverValid) {
                Spacer(Modifier.height(6.dp))
                Text("Cover image is required", fontSize = 12.sp, color = ErrorRed)
            }

            Spacer(modifier = Modifier.height(16.dp))

            // ─── Gallery ───
            Text(
                text = "Upload Gallery",
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
                color = TextDark,
                modifier = Modifier.padding(bottom = 8.dp),
            )
            GalleryUploadBox(
                images = galleryBytes,
                onAddClick = {
                    pickerTarget = "GALLERY"
                    showSourcePicker = true
                },
                onRemoveImage = { index ->
                    galleryBytes = galleryBytes.toMutableList().apply { removeAt(index) }
                },
            )

            Spacer(modifier = Modifier.height(16.dp))

            // ─── Email ───
            CustomTextField(
                label = "Email",
                placeholder = "e.g. 123@345",
                value = state.email,
                onValueChange = viewModel::onEmailChange,
                keyboardType = KeyboardType.Email,
                errorMessage = when {
                    !showErrors && state.email.isBlank() -> null
                    isEmailValid -> null
                    else -> "Enter a valid email (e.g. name@domain.com)"
                },
            )

            Spacer(modifier = Modifier.height(16.dp))

            // ─── Phone (digits only, max 10) ───
            CustomTextField(
                label = "Contact Number",
                placeholder = "e.g. 1234567890",
                value = state.contactNumber,
                onValueChange = { new ->
                    viewModel.onContactChange(new.filter { it.isDigit() }.take(10))
                },
                keyboardType = KeyboardType.Phone,
                errorMessage = when {
                    !showErrors && state.contactNumber.isBlank() -> null
                    isPhoneValid -> null
                    else -> "Enter a valid 10-digit phone number"
                },
            )

            Spacer(modifier = Modifier.height(24.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = "Allow Applications",
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Medium,
                    color = TextDark,
                )
                Switch(
                    checked = state.allowApplications,
                    onCheckedChange = viewModel::onAllowApplicationsChange,
                    colors = SwitchDefaults.colors(
                        checkedThumbColor = Color.White,
                        checkedTrackColor = PurplePrimary,
                        uncheckedThumbColor = Color.White,
                        uncheckedTrackColor = BorderGray,
                    ),
                )
            }

            Spacer(modifier = Modifier.height(80.dp))
        }
    }

    if (showSourcePicker) {
        ModalBottomSheet(
            onDismissRequest = { showSourcePicker = false },
            containerColor = Color.White,
            dragHandle = { BottomSheetDefaults.DragHandle() },
        ) {
            Column(Modifier.fillMaxWidth().padding(bottom = 40.dp, top = 8.dp)) {
                ListItem(
                    headlineContent = { Text("Gallery (Images only)") },
                    leadingContent = { Icon(Icons.Default.PhotoLibrary, null, tint = PurplePrimary) },
                    modifier = Modifier.clickable {
                        showSourcePicker = false
                        if (pickerTarget == "COVER") coverImagePicker.launch()
                        else galleryImagePicker.launch()
                    },
                )
                ListItem(
                    headlineContent = { Text("Files (All formats)") },
                    leadingContent = { Icon(Icons.Default.Folder, null, tint = PurplePrimary) },
                    modifier = Modifier.clickable {
                        showSourcePicker = false
                        if (pickerTarget == "COVER") coverFilePicker.launch()
                        else galleryFilePicker.launch()
                    },
                )
            }
        }
    }
}

// ─── Validated text field ───
@Composable
fun CustomTextField(
    label: String,
    placeholder: String,
    value: String,
    onValueChange: (String) -> Unit,
    trailingIcon: @Composable (() -> Unit)? = null,
    singleLine: Boolean = true,
    minHeight: Dp = 56.dp,
    keyboardType: KeyboardType = KeyboardType.Text,
    errorMessage: String? = null,
) {
    Column(modifier = Modifier.fillMaxWidth()) {
        Text(
            text = label,
            fontSize = 14.sp,
            fontWeight = FontWeight.Medium,
            color = TextDark,
            modifier = Modifier.padding(bottom = 8.dp),
        )
        OutlinedTextField(
            value = value,
            onValueChange = onValueChange,
            placeholder = {
                Text(text = placeholder, color = Color(0xFFBDBDBD), fontSize = 14.sp)
            },
            modifier = Modifier
                .fillMaxWidth()
                .defaultMinSize(minHeight = minHeight),
            shape = RoundedCornerShape(8.dp),
            colors = OutlinedTextFieldDefaults.colors(
                unfocusedBorderColor = if (errorMessage != null) ErrorRed else BorderGray,
                focusedBorderColor = if (errorMessage != null) ErrorRed else PurplePrimary,
                unfocusedContainerColor = Color.White,
                focusedContainerColor = Color.White,
            ),
            trailingIcon = trailingIcon,
            singleLine = singleLine,
            isError = errorMessage != null,
            keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
        )
        if (errorMessage != null) {
            Spacer(Modifier.height(4.dp))
            Text(errorMessage, fontSize = 12.sp, color = ErrorRed)
        }
    }
}

// ─── Cover box (16:9 preview) ───
@Composable
fun UploadBox(
    title: String,
    subtitle: String,
    previewBytes: ByteArray? = null,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(LightPurpleBg)
            .dashedBorder(color = PurplePrimary, strokeWidth = 1.5.dp, cornerRadius = 12.dp)
            .clickable { onClick() }
            .padding(if (previewBytes != null) 8.dp else 24.dp),
        contentAlignment = Alignment.Center,
    ) {
        if (previewBytes != null) {
            AsyncImage(
                model = previewBytes,
                contentDescription = "Cover Preview",
                contentScale = ContentScale.Crop,
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(16f / 9f)
                    .clip(RoundedCornerShape(12.dp)),
            )
        } else {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .clip(CircleShape)
                        .background(PurplePrimary),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        imageVector = Icons.Default.CloudUpload,
                        contentDescription = "Upload",
                        tint = Color.White,
                        modifier = Modifier.size(24.dp),
                    )
                }
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = title,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark,
                    textAlign = TextAlign.Center,
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = subtitle,
                    fontSize = 12.sp,
                    color = TextGray,
                    textAlign = TextAlign.Center,
                    lineHeight = 16.sp,
                )
            }
        }
    }
}

// ─── Gallery box (multiple images) ───
@OptIn(ExperimentalLayoutApi::class)
@Composable
fun GalleryUploadBox(
    images: List<ByteArray>,
    onAddClick: () -> Unit,
    onRemoveImage: (Int) -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(LightPurpleBg)
            .dashedBorder(color = PurplePrimary, strokeWidth = 1.5.dp, cornerRadius = 12.dp)
            .padding(12.dp),
    ) {
        if (images.isEmpty()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onAddClick() }
                    .padding(vertical = 12.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .clip(CircleShape)
                        .background(PurplePrimary),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Default.CloudUpload, "Upload", tint = Color.White, modifier = Modifier.size(24.dp))
                }
                Spacer(modifier = Modifier.height(12.dp))
                Text("Upload Gallery", fontSize = 16.sp, fontWeight = FontWeight.Bold, color = TextDark)
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    "Add multiple photos of your event",
                    fontSize = 12.sp,
                    color = TextGray,
                    textAlign = TextAlign.Center,
                )
            }
        } else {
            Column {
                Text(
                    "${images.size} photo${if (images.size > 1) "s" else ""} selected",
                    fontSize = 12.sp,
                    color = TextGray,
                    modifier = Modifier.padding(bottom = 8.dp),
                )
                FlowRow(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    images.forEachIndexed { index, bytes ->
                        Box(
                            modifier = Modifier
                                .size(80.dp)
                                .clip(RoundedCornerShape(8.dp)),
                        ) {
                            AsyncImage(
                                model = bytes,
                                contentDescription = null,
                                contentScale = ContentScale.Crop,
                                modifier = Modifier.fillMaxSize(),
                            )
                            Box(
                                modifier = Modifier
                                    .align(Alignment.TopEnd)
                                    .padding(4.dp)
                                    .size(20.dp)
                                    .clip(CircleShape)
                                    .background(Color.White)
                                    .clickable { onRemoveImage(index) },
                                contentAlignment = Alignment.Center,
                            ) {
                                Icon(
                                    Icons.Default.Close,
                                    contentDescription = "Remove",
                                    tint = Color.Red,
                                    modifier = Modifier.size(14.dp),
                                )
                            }
                        }
                    }
                    Box(
                        modifier = Modifier
                            .size(80.dp)
                            .clip(RoundedCornerShape(8.dp))
                            .background(Color.White)
                            .clickable { onAddClick() },
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(Icons.Default.Add, "Add more", tint = PurplePrimary)
                    }
                }
            }
        }
    }
}