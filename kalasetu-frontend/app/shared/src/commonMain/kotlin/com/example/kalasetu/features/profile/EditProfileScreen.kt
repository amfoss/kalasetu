package com.example.kalasetu.features.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CameraAlt
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Email
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.PhotoLibrary
import androidx.compose.material.icons.filled.Tag
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

private val ErrorRed      = Color(0xFFB00020)



data class EditProfileState(
    val name: String = "",
    val location: String = "",
    val bio: String = "",
    val email: String = "",
    val avatarBytes: ByteArray? = null,
    val nameError: String? = null,
    val emailError: String? = null,
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other == null || (this::class != other::class)) return false

        other as EditProfileState

        if (name != other.name) return false
        if (location != other.location) return false
        if (bio != other.bio) return false
        if (email != other.email) return false
        if (avatarBytes != null) {
            if (other.avatarBytes == null) return false
            if (!avatarBytes.contentEquals(other.avatarBytes)) return false
        } else if (other.avatarBytes != null) return false
        if (nameError != other.nameError) return false
        return emailError == other.emailError
    }

    override fun hashCode(): Int {
        var result = name.hashCode()
        result = 31 * result + location.hashCode()
        result = 31 * result + bio.hashCode()
        result = 31 * result + email.hashCode()
        result = 31 * result + (avatarBytes?.contentHashCode() ?: 0)
        result = 31 * result + (nameError?.hashCode() ?: 0)
        result = 31 * result + (emailError?.hashCode() ?: 0)
        return result
    }
}



@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EditProfileScreen(
    profile: Profile,
    onBack: () -> Unit = {},
    onSave: (Profile) -> Unit = {},
) {
    var state by remember {
        mutableStateOf(
            EditProfileState(
                name     = profile.name,
                location = profile.location,
                bio      = profile.bio,
                email    = profile.email,
                avatarBytes = profile.avatarBytes
            )
        )
    }

    var showImageSourceOption by remember { mutableStateOf(value = false) }

    val imagePickerLauncher = rememberImagePicker { bytes ->
        bytes?.let {
            state = state.copy(avatarBytes = it)
        }
    }

    fun validate(): Boolean {
        val nameErr = if (state.name.isBlank()) "Name cannot be empty" else null
        val emailErr = when {
            state.email.isBlank()        -> "Email cannot be empty"
            !state.email.contains("@")   -> "Enter a valid email"
            else                         -> null
        }
        state = state.copy(
            nameError     = nameErr,
            emailError    = emailErr
        )
        return (nameErr == null) && (emailErr == null)
    }

    fun buildUpdatedProfile() = profile.copy(
        name     = state.name.trim(),
        location = state.location.trim(),
        bio      = state.bio.trim(),
        email    = state.email.trim(),
        avatarBytes = state.avatarBytes
    )

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.surface)
            .imePadding()
    ) {
        // Header

        EditProfileHeader(
            initials = state.name.toInitials().ifBlank { profile.name.toInitials() },
            avatarBytes = state.avatarBytes,
            onBack   = onBack,
            onCameraClick = { showImageSourceOption = true },
            onSave   = { if (validate()) onSave(buildUpdatedProfile()) }
        )

        //  Scrollable form
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 20.dp)
                .padding(top = 28.dp, bottom = 40.dp),
            verticalArrangement = Arrangement.spacedBy(20.dp)
        ) {
            SectionLabel("Personal Information")

            EditField(
                label         = "Full Name",
                value         = state.name,
                placeholder   = "Enter your full name",
                icon          = Icons.Default.Person,
                error         = state.nameError,
            ) { state = state.copy(name = it, nameError = null) }

            EditField(
                label         = "Location",
                value         = state.location,
                placeholder   = "Enter your city",
                icon          = Icons.Default.LocationOn,
                onValueChange = { state = state.copy(location = it) }
            )

            SectionLabel("About")

            BioField(
                value         = state.bio,
            ) { state = state.copy(bio = it) }

            SectionLabel("Contact")

            EditField(
                label         = "Email",
                value         = state.email,
                placeholder   = "Enter your email",
                icon          = Icons.Default.Email,
                error         = state.emailError,
                onValueChange = { state = state.copy(email = it, emailError = null) }
            )

            Spacer(Modifier.height(4.dp))

            // Bottom save button
            Button(
                onClick  = { if (validate()) onSave(buildUpdatedProfile()) },
                modifier = Modifier.fillMaxWidth().height(52.dp),
                shape    = RoundedCornerShape(14.dp),
                colors   = ButtonDefaults.buttonColors(containerColor = BrandPurple)
            ) {
                Icon(
                    imageVector        = Icons.Default.Check,
                    contentDescription = null,
                    modifier           = Modifier.size(18.dp)
                )
                Spacer(Modifier.width(8.dp))
                Text(
                    text       = "Save Changes",
                    fontSize   = 16.sp,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }

    if (showImageSourceOption) {
        ModalBottomSheet(
            onDismissRequest = { showImageSourceOption = false },
            dragHandle = { BottomSheetDefaults.DragHandle() },
            containerColor = MaterialTheme.colorScheme.surface,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 32.dp, top = 8.dp)
            ) {
                ListItem(
                    headlineContent = { Text("Choose from gallery") },
                    leadingContent = { Icon(Icons.Default.PhotoLibrary, null) },
                    modifier = Modifier.clickable {
                        showImageSourceOption = false
                        imagePickerLauncher.launch()
                    }
                )
                if (state.avatarBytes != null) {
                    ListItem(
                        headlineContent = { Text("Remove profile picture", color = ErrorRed) },
                        leadingContent = { Icon(Icons.Default.Delete, null, tint = ErrorRed) },
                        modifier = Modifier.clickable {
                            showImageSourceOption = false
                            state = state.copy(avatarBytes = null)
                        }
                    )
                }
            }
        }
    }
}

// Header

@Composable
private fun EditProfileHeader(
    initials: String,
    avatarBytes: ByteArray?,
    onBack: () -> Unit,
    onCameraClick: () -> Unit,
    onSave: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(200.dp),
    ) {
        // Gradient banner
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(150.dp)
                .background(
                    brush = Brush.linearGradient(listOf(BrandPurple, LightPurple))
                )
        ) {

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(
                        top = 48.dp,
                        start = 8.dp,
                        end = 8.dp
                    ),
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Back
                IconButton(
                    onClick = onBack,
                    modifier = Modifier.size(48.dp)
                ) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                        contentDescription = "Back",
                        tint = Color.White
                    )
                }

                // Title
                Box(
                    modifier = Modifier.weight(1f),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "Edit Profile",
                        fontSize = 18.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = Color.White
                    )
                }

                // Save
                Text(
                    text = "Save",
                    fontSize = 15.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White,
                    modifier = Modifier
                        .clip(RoundedCornerShape(8.dp))
                        .clickable(onClick = onSave)
                        .padding(
                            horizontal = 12.dp,
                            vertical = 12.dp
                        )
                )
            }
        }

        // Avatar
        Box(modifier = Modifier.align(Alignment.BottomCenter)) {
            Box(
                modifier = Modifier
                    .size(90.dp)
                    .clip(CircleShape)
                    .border(3.dp, MaterialTheme.colorScheme.surface, CircleShape)
                    .background(LightPurple),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text       = initials,
                    color      = Color.White,
                    fontSize   = 28.sp,
                    fontWeight = FontWeight.Bold
                )

                avatarBytes?.let { bytes ->
                    coil3.compose.AsyncImage(
                        model = bytes,
                        contentDescription = null,
                        modifier = Modifier.fillMaxSize().clip(CircleShape),
                        contentScale = androidx.compose.ui.layout.ContentScale.Crop
                    )
                }
            }

            // Camera badge — bottom-right of avatar
            Box(
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .size(28.dp)
                    .clip(CircleShape)
                    .background(BrandPurple)
                    .border(2.dp, MaterialTheme.colorScheme.surface, CircleShape)
                    .clickable(onClick = onCameraClick),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector        = Icons.Default.CameraAlt,
                    contentDescription = "Change photo",
                    tint               = Color.White,
                    modifier           = Modifier.size(14.dp)
                )
            }
        }
    }
}

@Composable
private fun SectionLabel(text: String) {
    Text(
        text          = text,
        fontSize      = 12.sp,
        fontWeight    = FontWeight.Bold,
        color         = BrandPurple,
        letterSpacing = 0.8.sp,
    )
}

@Composable
private fun EditField(
    label: String,
    value: String,
    placeholder: String,
    icon: ImageVector,
    error: String? = null,
    prefix: String? = null,
    onValueChange: (String) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(
            text       = label,
            fontSize   = 13.sp,
            fontWeight = FontWeight.Medium,
            color      = MaterialTheme.colorScheme.onSurface
        )
        OutlinedTextField(
            value         = value,
            onValueChange = onValueChange,
            placeholder   = {
                Text(text = placeholder, color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 14.sp)
            },
            leadingIcon   = {
                Icon(
                    imageVector        = icon,
                    contentDescription = null,
                    tint               = if (error != null) ErrorRed else BrandPurple,
                    modifier           = Modifier.size(18.dp)
                )
            },
            prefix = prefix?.let { p ->
                { Text(text = p, color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 14.sp) }
            },
            isError    = error != null,
            singleLine = true,
            shape      = RoundedCornerShape(12.dp),
            modifier   = Modifier.fillMaxWidth(),
            colors     = OutlinedTextFieldDefaults.colors(
                focusedBorderColor      = BrandPurple,
                unfocusedBorderColor    = MaterialTheme.colorScheme.outlineVariant,
                errorBorderColor        = ErrorRed,
                focusedContainerColor   = MaterialTheme.colorScheme.surfaceVariant,
                unfocusedContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                errorContainerColor     = MaterialTheme.colorScheme.surfaceVariant
            )
        )
        error?.let { err ->
            Text(text = err, fontSize = 12.sp, color = ErrorRed)
        }
    }
}

@Composable
private fun BioField(
    value: String,
    onValueChange: (String) -> Unit,
) {
    val maxChars = 200
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(
            text       = "Bio",
            fontSize   = 13.sp,
            fontWeight = FontWeight.Medium,
            color      = MaterialTheme.colorScheme.onSurface
        )
        OutlinedTextField(
            value         = value,
            onValueChange = { if (it.length <= maxChars) onValueChange(it) },
            placeholder   = {
                Text(
                    text     = "Tell the world about your art and what inspires you…",
                    color    = MaterialTheme.colorScheme.onSurfaceVariant,
                    fontSize = 14.sp
                )
            },
            minLines = 4,
            maxLines = 6,
            shape    = RoundedCornerShape(12.dp),
            modifier = Modifier.fillMaxWidth(),
            colors   = OutlinedTextFieldDefaults.colors(
                focusedBorderColor      = BrandPurple,
                unfocusedBorderColor    = MaterialTheme.colorScheme.outlineVariant,
                focusedContainerColor   = MaterialTheme.colorScheme.surfaceVariant,
                unfocusedContainerColor = MaterialTheme.colorScheme.surfaceVariant
            )
        )
        Text(
            text     = "${value.length} / $maxChars",
            fontSize = 11.sp,
            color    = if (value.length >= maxChars) ErrorRed else MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.align(Alignment.End)
        )
    }
}
