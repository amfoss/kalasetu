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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun ProfileHeader(
    profile: Profile,
    onEditProfile: () -> Unit,
    onShare: () -> Unit,
    onBack: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(160.dp),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(110.dp)
                .background(LightPurple
                    ),
        )

        IconButton(
            onClick = onBack,
            modifier = Modifier
                .align(Alignment.TopStart)
                .padding(top = 40.dp, start = 8.dp)
        ) {
            Icon(
                imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                contentDescription = "Back",
                tint = Color.White
            )
        }

        Box(
            modifier = Modifier
                .align(Alignment.BottomStart)
                .padding(start = 16.dp),
        ) {
            ProfileAvatar(
                initials = profile.name.toInitials(),
                imageUrl = profile.avatarUrl,
                avatarBytes = profile.avatarBytes,
            )
        }

        Row(
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .padding(end = 16.dp, bottom = 2.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(CircleShape)
                    .background(CardWhite)
                    .border(1.dp, DividerGray, CircleShape)
                    .clickable(onClick = onShare),
                contentAlignment = Alignment.Center,
            ) {
                Icon(
                    imageVector = Icons.Default.Share,
                    contentDescription = "Share profile",
                    tint = TextSecondary,
                    modifier = Modifier.size(18.dp),
                )
            }

            Button(
                onClick = onEditProfile,
                shape = RoundedCornerShape(20.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = BrandPurple,
                ),
            ) {
                Icon(
                    imageVector = Icons.Default.Edit,
                    contentDescription = null,
                    modifier = Modifier.size(14.dp),
                )

                Spacer(Modifier.width(6.dp))

                Text(
                    "Edit Profile",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Medium,
                )
            }
        }
    }
}

@Composable
fun ProfileAvatar(
    initials: String,
    imageUrl: String?,
    avatarBytes: ByteArray? = null,
) {
    Box(
        modifier = Modifier
            .size(80.dp)
            .clip(CircleShape)
            .border(3.dp, CardWhite, CircleShape)
            .background(LightPurple),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = initials,
            color = Color.White,
            fontSize = 24.sp,
            fontWeight = FontWeight.Bold,
        )

        avatarBytes?.let { bytes ->
            coil3.compose.AsyncImage(
                model = bytes,
                contentDescription = null,
                modifier = Modifier.fillMaxSize().clip(CircleShape),
                contentScale = androidx.compose.ui.layout.ContentScale.Crop,
            )
        } ?: imageUrl?.let { url ->
            coil3.compose.AsyncImage(
                model = url,
                contentDescription = null,
                modifier = Modifier.fillMaxSize().clip(CircleShape),
                contentScale = androidx.compose.ui.layout.ContentScale.Crop,
            )
        }
    }
}

@Composable
fun ProfileInfo(profile: Profile) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 20.dp)
            .padding(top = 12.dp),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(
                text = profile.name,
                fontSize = 22.sp,
                fontWeight = FontWeight.Bold,
                color = TextPrimary,
            )

            if (profile.isVerified) {
                Spacer(Modifier.width(6.dp))

                Icon(
                    imageVector = Icons.Default.CheckCircle,
                    contentDescription = "Verified",
                    tint = BrandPurple,
                    modifier = Modifier.size(20.dp),
                )
            }
        }

        Spacer(Modifier.height(2.dp))

        if (profile.username.isNotBlank()) {
            Text(
                text = "@${profile.username}",
                fontSize = 14.sp,
                color = TextSecondary,
            )
        }

        if (profile.location.isNotBlank()) {
            Spacer(Modifier.height(6.dp))

            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    imageVector = Icons.Default.LocationOn,
                    contentDescription = "Location",
                    tint = TextSecondary,
                    modifier = Modifier.size(14.dp),
                )

                Spacer(Modifier.width(4.dp))

                Text(
                    text = profile.location,
                    fontSize = 13.sp,
                    color = TextSecondary,
                )
            }
        }

        Spacer(Modifier.height(10.dp))

        if (profile.bio.isNotBlank()) {
            Text(
                text = profile.bio,
                fontSize = 14.sp,
                color = TextPrimary,
            )
        }

        Spacer(Modifier.height(16.dp))

        Row(
            horizontalArrangement = Arrangement.spacedBy(20.dp),
        ) {
            InlineStat(
                value = profile.followers.toDisplayCount(),
                label = "Followers",
            )

            InlineStat(
                value = profile.following.toDisplayCount(),
                label = "Following",
            )

            InlineStat(
                value = profile.artworksCount.toDisplayCount(),
                label = "Artworks",
            )
        }
    }
}

@Composable
fun InlineStat(
    value: String,
    label: String,
) {
    Column {
        Text(
            text = value,
            fontSize = 18.sp,
            fontWeight = FontWeight.Bold,
            color = TextPrimary,
        )

        Text(
            text = label,
            fontSize = 12.sp,
            color = TextSecondary,
        )
    }
}