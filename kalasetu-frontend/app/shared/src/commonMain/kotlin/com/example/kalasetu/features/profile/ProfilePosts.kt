package com.example.kalasetu.features.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.wrapContentHeight
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Share
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.ThumbUp
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import coil3.compose.AsyncImage
import kotlinx.datetime.Instant
import kotlin.time.Clock

object RelativeTime {
    fun format(createdAt: String): String = try {
        val diffMinutes = (Clock.System.now().toEpochMilliseconds() - Instant.parse(createdAt).toEpochMilliseconds()) / 60_000L
        when {
            diffMinutes < 1 -> "Just now"
            diffMinutes < 60 -> "${diffMinutes}m ago"
            diffMinutes < 24 * 60 -> "${diffMinutes / 60}h ago"
            else -> "${diffMinutes / (24 * 60)}d ago"
        }
    } catch (e: Exception) {
        "Just now"
    }
}

data class DraftPost(
    val timeAgo: String,
    val content: String,
    val likes: Int,
    val comments: Int,
    val hasImage: Boolean,
    val imageBytes: List<ByteArray> = emptyList(),
    val imageUrl: String? = null,
)

internal fun profilePostToDraftPost(post: ProfilePost): DraftPost = DraftPost(
    timeAgo = RelativeTime.format(post.createdAt),
    content = post.content,
    likes = post.likeCount,
    comments = post.commentCount,
    hasImage = post.mediaUri != null,
    imageUrl = post.mediaUri
)

@Composable
fun PostsTabContent(
    profile: Profile,
    posts: List<DraftPost>
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        posts.forEach { post ->
            PostCard(
                profile = profile,
                post = post
            )
        }
    }
}

@Composable
fun PostCard(
    profile: Profile,
    post: DraftPost
) {
    var fullScreenImageIndex by remember { mutableStateOf<Int?>(null) }
    val imagesToDisplay = if (post.imageUrl != null) listOf(post.imageUrl) else post.imageBytes

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = CardDefaults.cardElevation(
            defaultElevation = 1.dp
        )
    ) {
        Column(
            modifier = Modifier.fillMaxWidth()
        ) {

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(
                        horizontal = 12.dp,
                        vertical = 10.dp
                    ),
                verticalAlignment = Alignment.CenterVertically
            ) {
                ProfileAvatar(
                    initials = profile.name.toInitials(),
                    imageUrl = profile.avatarUrl,
                    avatarBytes = profile.avatarBytes
                )

                Spacer(Modifier.width(10.dp))

                Column(
                    modifier = Modifier.weight(1f)
                ) {
                    Text(
                        text = profile.name,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = MaterialTheme.colorScheme.onSurface
                    )

                    Text(
                        text = post.timeAgo,
                        fontSize = 12.sp,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            Text(
                text = post.content,
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onSurface,
                maxLines = 4,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.padding(horizontal = 14.dp)
            )

            val displayImages = if (post.imageUrl != null) listOf(post.imageUrl) else post.imageBytes
            if (displayImages.isNotEmpty()) {
                Spacer(Modifier.height(10.dp))
                PostImageCarousel(
                    images = displayImages,
                    onImageClick = { index -> fullScreenImageIndex = index }
                )
            }

            Spacer(Modifier.height(10.dp))

            Row(
                modifier = Modifier.padding(horizontal = 14.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = Icons.Default.Favorite,
                    contentDescription = null,
                    tint = BrandPurple,
                    modifier = Modifier.size(14.dp)
                )

                Text(
                    text = "${post.likes} likes",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Text("•")

                Text(
                    text = "${post.comments} comments",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            Spacer(Modifier.height(8.dp))

            HorizontalDivider(
                color = MaterialTheme.colorScheme.outlineVariant,
                thickness = 1.dp
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(
                        horizontal = 8.dp,
                        vertical = 4.dp
                    ),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                PostActionButton(
                    icon = Icons.Default.ThumbUp,
                    label = "Like"
                )

                PostActionButton(
                    icon = Icons.Default.Star,
                    label = "Comment"
                )

                PostActionButton(
                    icon = Icons.Default.Share,
                    label = "Share"
                )
            }
        }
    }

    if (fullScreenImageIndex != null) {
        FullScreenImageViewer(
            images = imagesToDisplay,
            initialPage = fullScreenImageIndex!!,
            onDismiss = { fullScreenImageIndex = null }
        )
    }
}

@Composable
internal fun PostImageCarousel(
    images: List<Any>,
    onImageClick: (Int) -> Unit
) {
    var maxHeightPx by remember { mutableIntStateOf(0) }
    val density = LocalDensity.current

    BoxWithConstraints(
        modifier = Modifier
            .fillMaxWidth()
            .background(Purple50.copy(alpha = 0.5f))
    ) {
        val maxWidthPx = constraints.maxWidth

        val heightModifier = if (maxHeightPx > 0) {
            Modifier.height(with(density) { maxHeightPx.toDp() })
        } else {
            Modifier.heightIn(min = 200.dp, max = 500.dp).wrapContentHeight()
        }

        val pagerState = rememberPagerState(pageCount = { images.size })

        HorizontalPager(
            state = pagerState,
            modifier = heightModifier.fillMaxWidth()
        ) { page ->
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .clickable { onImageClick(page) },
                contentAlignment = Alignment.Center
            ) {
                AsyncImage(
                    model = images[page],
                    contentDescription = null,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier.fillMaxSize(),
                    onSuccess = { state ->
                        if (page == 0 && maxHeightPx == 0) {
                            val intrinsicSize = state.painter.intrinsicSize
                            if (intrinsicSize.width > 0) {
                                val aspect = intrinsicSize.height / intrinsicSize.width
                                val clampedAspect = aspect.coerceIn(0.5f, 1.25f)
                                maxHeightPx = (maxWidthPx * clampedAspect).toInt()
                            }
                        }
                    }
                )
            }
        }

        if (images.size > 1) {
            Row(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 12.dp)
                    .background(Color.Black.copy(alpha = 0.3f), RoundedCornerShape(12.dp))
                    .padding(horizontal = 8.dp, vertical = 4.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                repeat(images.size) { index ->
                    Box(
                        modifier = Modifier
                            .size(6.dp)
                            .clip(CircleShape)
                            .background(
                                if (index == pagerState.currentPage)
                                    Color.White
                                else
                                    Color.White.copy(alpha = 0.5f)
                            )
                    )
                }
            }
        }
    }
}

@Composable
fun FullScreenImageViewer(
    images: List<Any>,
    initialPage: Int,
    onDismiss: () -> Unit
) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black)
        ) {
            val pagerState = rememberPagerState(initialPage = initialPage, pageCount = { images.size })

            HorizontalPager(
                state = pagerState,
                modifier = Modifier.fillMaxSize(),
                pageSpacing = 16.dp
            ) { page ->
                AsyncImage(
                    model = images[page],
                    contentDescription = null,
                    contentScale = ContentScale.Fit,
                    modifier = Modifier.fillMaxSize()
                )
            }

            IconButton(
                onClick = onDismiss,
                modifier = Modifier
                    .align(Alignment.TopStart)
                    .padding(top = 40.dp, start = 16.dp)
            ) {
                Icon(Icons.Default.Close, contentDescription = "Close", tint = Color.White)
            }

            if (images.size > 1) {
                Text(
                    text = "${pagerState.currentPage + 1} / ${images.size}",
                    color = Color.White,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 40.dp),
                    fontSize = 16.sp
                )
            }
        }
    }
}

@Composable
fun PostActionButton(
    icon: ImageVector,
    label: String
) {
    Row(
        modifier = Modifier
            .clip(RoundedCornerShape(8.dp))
            .clickable { }
            .padding(
                horizontal = 12.dp,
                vertical = 8.dp
            ),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(6.dp)
    ) {
        Icon(
            imageVector = icon,
            contentDescription = label,
            tint = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.size(18.dp)
        )

        Text(
            text = label,
            fontSize = 13.sp,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            fontWeight = FontWeight.Medium
        )
    }
}