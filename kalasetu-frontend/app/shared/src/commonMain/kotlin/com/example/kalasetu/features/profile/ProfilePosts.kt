package com.example.kalasetu.features.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Share
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.ThumbUp
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.layout.ContentScale
import coil3.compose.AsyncImage
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.layout.fillMaxSize
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
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = CardWhite
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
                        color = TextPrimary
                    )

                    Text(
                        text = post.timeAgo,
                        fontSize = 12.sp,
                        color = TextSecondary
                    )
                }
            }

            Text(
                text = post.content,
                fontSize = 14.sp,
                color = TextPrimary,
                maxLines = 4,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.padding(horizontal = 14.dp)
            )

            if (post.imageUrl != null) {
                Spacer(Modifier.height(10.dp))

                AsyncImage(
                    model = post.imageUrl,
                    contentDescription = "Post image",
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(300.dp)
                )
            } else if (post.imageBytes.isNotEmpty()) {
                Spacer(Modifier.height(10.dp))

                PostImageCarousel(
                    imageBytes = post.imageBytes
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
                    color = TextSecondary
                )

                Text("•")

                Text(
                    text = "${post.comments} comments",
                    fontSize = 12.sp,
                    color = TextSecondary
                )
            }

            Spacer(Modifier.height(8.dp))

            HorizontalDivider(
                color = DividerGray,
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
}
@Composable
private fun PostImageCarousel(
    imageBytes: List<ByteArray>
) {
    val pagerState = rememberPagerState(
        pageCount = { imageBytes.size }
    )

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(300.dp)
    ) {
        HorizontalPager(
            state = pagerState,
            modifier = Modifier.fillMaxSize()
        ) { page ->

            AsyncImage(
                model = imageBytes[page],
                contentDescription = "Post image ${page + 1}",
                contentScale = ContentScale.Fit,
                modifier = Modifier.fillMaxSize()
            )
        }

        // Page indicators
        if (imageBytes.size > 1) {
            Row(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                repeat(imageBytes.size) { index ->
                    Box(
                        modifier = Modifier
                            .size(6.dp)
                            .clip(androidx.compose.foundation.shape.CircleShape)
                            .background(
                                if (index == pagerState.currentPage)
                                    BrandPurple
                                else
                                    androidx.compose.ui.graphics.Color.Gray.copy(
                                        alpha = 0.6f
                                    )
                            )
                    )
                }
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
            tint = TextSecondary,
            modifier = Modifier.size(18.dp)
        )

        Text(
            text = label,
            fontSize = 13.sp,
            color = TextSecondary,
            fontWeight = FontWeight.Medium
        )
    }
}