package com.example.kalasetu.features.feed

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.HelpOutline
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil3.compose.AsyncImage
import com.example.kalasetu.features.profile.toInitials
import com.example.kalasetu.theme.SelectedPurple

@Composable
fun SidebarContent(
    userName: String?,
    userEmail: String?,
    userAvatarUrl: String?,
    userAvatarBytes: ByteArray?,
    currentRoute: String,
    onClose: () -> Unit,
    onNavigate: (String) -> Unit
) {
    ModalDrawerSheet(
        drawerContainerColor = Color.White,
        drawerShape = RoundedCornerShape(topEnd = 16.dp, bottomEnd = 16.dp),
        modifier = Modifier.fillMaxHeight().width(300.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(20.dp)
                .verticalScroll(rememberScrollState())
        ) {

            Spacer(modifier = Modifier.height(16.dp))

            // User Profile Header
            Row(verticalAlignment = Alignment.CenterVertically) {
                val model = userAvatarBytes ?: userAvatarUrl
                Box(
                    modifier = Modifier
                        .size(60.dp)
                        .clip(CircleShape)
                        .background(MaterialTheme.colorScheme.primaryContainer),
                    contentAlignment = Alignment.Center
                ) {
                    if (model != null) {
                        AsyncImage(
                            model = model,
                            contentDescription = "Profile",
                            contentScale = ContentScale.Crop,
                            modifier = Modifier.fillMaxSize()
                        )
                    } else {
                        val initials = userName?.toInitials() ?: ""
                        if (initials.isNotEmpty()) {
                            Text(
                                text = initials,
                                color = MaterialTheme.colorScheme.onPrimaryContainer,
                                fontWeight = FontWeight.Bold,
                                fontSize = 24.sp
                            )
                        } else {
                            Icon(
                                imageVector = Icons.Default.Person,
                                contentDescription = "Profile",
                                tint = MaterialTheme.colorScheme.onSurfaceVariant,
                                modifier = Modifier.size(32.dp)
                            )
                        }
                    }
                }
                Spacer(modifier = Modifier.width(12.dp))
                Column {
                    Text(
                        text = userName ?: "John Doe",
                        fontWeight = FontWeight.Bold,
                        fontSize = 20.sp
                    )
                    Text(
                        text = userEmail ?: "",
                        color = Color.Gray,
                        fontSize = 11.sp
                    )
                }
            }

            Spacer(modifier = Modifier.height(32.dp))

            // Menu Items
            SidebarItem(
                icon = Icons.Default.GridView,
                label = "Dashboard",
                isSelected = currentRoute == "Dashboard",
                onClick = { onNavigate("Dashboard") }
            )

            SectionLabel("Manage")
            SidebarItem(icon = Icons.Outlined.AssignmentInd, label = "Applications", isSelected = currentRoute == "Applications",onClick = { onNavigate("Applications") })
            SidebarItem(icon = Icons.Outlined.ConfirmationNumber, label = "My Events",isSelected = currentRoute == "MyEvents", onClick = { onNavigate("MyEvents") })

            //HorizontalDivider(modifier = Modifier.padding(vertical = 16.dp), color = Color(0xFFEEEEEE))

            //SectionLabel("Communication")
            //SidebarItem(icon = Icons.Outlined.Notifications, label = "Notifications",isSelected = currentRoute == "Notifications", onClick = { onNavigate("Notifications") })
            //SidebarItem(icon = Icons.Outlined.Campaign, label = "Announcements",isSelected = currentRoute == "Announcements", onClick = { onNavigate("Announcements") })

            HorizontalDivider(modifier = Modifier.padding(vertical = 16.dp), color = Color(0xFFEEEEEE))

            SectionLabel("Profile")
            SidebarItem(icon = Icons.Outlined.Settings, label = "Settings",isSelected = currentRoute == "Settings", onClick = { onNavigate("Settings") })
            SidebarItem(icon = Icons.Outlined.Person, label = "Profile",isSelected = currentRoute == "Profile", onClick = { onNavigate("Profile") })
            SidebarItem(icon = Icons.Outlined.Info, label = "Help ?",isSelected = currentRoute == "Help", onClick = { onNavigate("Help") })
        }
    }
}

@Composable
private fun SectionLabel(text: String) {
    Text(
        text = text,
        color = Color.Gray,
        fontSize = 12.sp,
        modifier = Modifier.padding(vertical = 8.dp, horizontal = 4.dp)
    )
}

@Composable
private fun SidebarItem(
    icon: ImageVector,
    label: String,
    isSelected: Boolean = false,
    onClick: () -> Unit
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp)
            .clickable { onClick() },
        color = if (isSelected) SelectedPurple.copy(alpha = 0.15f) else Color.Transparent,
        shape = RoundedCornerShape(8.dp)
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = if (isSelected) SelectedPurple else SelectedPurple.copy(alpha = 0.8f),
                modifier = Modifier.size(22.dp)
            )
            Spacer(modifier = Modifier.width(16.dp))
            Text(
                text = label,
                color = if (isSelected) SelectedPurple else Color.DarkGray,
                fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Medium,
                fontSize = 15.sp
            )
        }
    }
}
