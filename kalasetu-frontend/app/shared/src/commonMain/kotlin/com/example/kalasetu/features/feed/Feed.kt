package com.example.kalasetu.features.feed
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.launch
import coil3.compose.AsyncImage
import com.example.kalasetu.features.profile.*
import com.example.kalasetu.features.profile.toInitials
import com.example.kalasetu.features.settings.SavedStore
import com.example.kalasetu.theme.KalasetuTheme
import com.example.kalasetu.theme.SelectedPurple
import com.example.kalasetu.theme.SubtitleGray
import com.example.kalasetu.theme.UnselectedBorder
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties

data class ArtistPost(
    val id: Int,
    val userId: Int = 0,
    val artistName: String,
    val craft: String,
    val location: String,
    val timeAgo: String,
    val avatarUrl: String,
    val images: List<String>,
    val caption: String,
    val likes: Int,
    val comments: Int,
    val isLiked: Boolean = false,
    val isMine: Boolean = false,
    val avatarBackground: Color = Color(0xFFEDE7F6)
)

//Hardcoded stuff
private val dummyPosts = listOf(
    ArtistPost(
        id = 1,
        artistName = "Praharsha",
        craft = "Artist",
        location = "India",
        timeAgo = "2h ago",
        avatarUrl = "https://i.pravatar.cc/150?img=12",
        images = listOf(
            "https://picsum.photos/seed/kasavu1/800/600",
            "https://picsum.photos/seed/kasavu2/800/600",
            "https://picsum.photos/seed/kasavu3/800/600"
        ),
        caption = "Love making portraits and also do enjoy acrylic painting...",
        likes = 10006,
        comments = 45
    ),
    ArtistPost(
        id = 2,
        artistName = "nkart826",
        craft = "Charcoal Painting",
        location = "India",
        timeAgo = "4w ago",
        avatarUrl = "https://i.pravatar.cc/150?img=33",
        images = listOf(
            "https://picsum.photos/seed/horse1/800/900"
        ),
        caption = "Studies in charcoal — capturing motion and light in monochrome.",
        likes = 512,
        comments = 87
    )
)


@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedScreen(
    viewModel: FeedViewModel,
    userAvatarUrl: String? = null,
    userAvatarBytes: ByteArray? = null,
    userName: String? = null,
    onNavigateToProfile: () -> Unit = {},
    onNavigateToStore: () -> Unit = {},
    onNavigateToEvents: () -> Unit = {},
    onNavigateToHome: () -> Unit = {},
    onMenuClick: () -> Unit = {}
) {
    val posts by viewModel.posts.collectAsState()
    val commentsList by viewModel.comments.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    val savedPosts by SavedStore.saved.collectAsState()
    val savedIds = remember(savedPosts) { savedPosts.map { it.id }.toSet() }

    LaunchedEffect(Unit) {
        viewModel.loadPosts()
    }
    var selectedTab by remember { mutableIntStateOf(0) }
    val tabs = listOf("FEED", "DISCOVER")
    var showComments by remember { mutableStateOf(false) }
    var activePostIdForComments by remember { mutableStateOf<Int?>(null) }

    val commentsSheetState = rememberModalBottomSheetState(
        skipPartiallyExpanded = true
    )

    Scaffold(
        containerColor = MaterialTheme.colorScheme.background,
        topBar = {
            KalaTopBar(
                avatarUrl = userAvatarUrl,
                avatarBytes = userAvatarBytes,
                userName = userName,
                onProfileClick = onNavigateToProfile,
                onMenuClick = onMenuClick
            )
        },
        bottomBar = {
            KalaBottomNav(
                selectedIndex = 0, // Home/Dashboard
                onStoreClick = onNavigateToStore,
                onEventsClick = onNavigateToEvents,
                onHomeClick = onNavigateToHome,
                onProfileClick = onNavigateToProfile
            )
        }
    ) { innerPadding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
        ) {
            KalaTabRow(
                tabs = tabs,
                selectedIndex = selectedTab,
                onTabSelected = { selectedTab = it }
            )

            when (selectedTab) {
                0 -> FeedContent(
                    posts = posts,
                    savedIds = savedIds,
                    onSaveToggle = { post -> SavedStore.toggle(post) },
                    onLikeClick = { postId -> viewModel.toggleLike(postId) },
                    onDeleteClick = { postId -> viewModel.deletePost(postId) },
                    onCommentClick = { postId ->
                        activePostIdForComments = postId
                        viewModel.loadComments(postId)
                        showComments = true
                    }
                )
                1 -> DiscoverContent()
                else -> {
                    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Text("Content for ${tabs[selectedTab]}")
                    }
                }
            }
        }

        if (showComments && activePostIdForComments != null) {
            CommentsBottomSheet(
                postId = activePostIdForComments!!,
                commentsList = commentsList,
                onSendComment = { commentText ->
                    viewModel.addComment(activePostIdForComments!!, commentText)
                },
                onDismissRequest = {
                    showComments = false
                },
                sheetState = commentsSheetState
            )
        }
    }
}

@Composable
internal fun FeedContent(
    posts: List<ArtistPost>,
    savedIds: Set<Int>,
    onSaveToggle: (ArtistPost) -> Unit,
    onLikeClick: (Int) -> Unit,
    onDeleteClick: (Int) -> Unit,
    onCommentClick: (Int) -> Unit
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(vertical = 8.dp)
    ) {
        items(posts) { post ->
            PostCard(
                post = post,
                saved = post.id in savedIds,
                onSaveClick = { onSaveToggle(post) },
                onLikeClick = { onLikeClick(post.id) },
                onDeleteClick = { onDeleteClick(post.id) },
                onCommentClick = { onCommentClick(post.id) }
            )
            Spacer(modifier = Modifier.height(8.dp))
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun KalaTopBar(
    avatarUrl: String?,
    avatarBytes: ByteArray?,
    userName: String?,
    onProfileClick: () -> Unit,
    onMenuClick: () -> Unit,
    title: String = "KalaSetu",
    onBack: (() -> Unit)? = null,
) {
    CenterAlignedTopAppBar(
        title = {
            Text(
                text = title,
                fontWeight = FontWeight.Bold,
                fontSize = 22.sp
            )
        },
        navigationIcon = {
            if (onBack != null) {
                IconButton(onClick = onBack) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                        contentDescription = "Back"
                    )
                }
            } else {
                IconButton(onClick = onMenuClick) {
                    Icon(Icons.Default.Menu, contentDescription = "Menu")
                }
            }
        },
        actions = {
            val model = avatarBytes ?: avatarUrl
            Box(
                modifier = Modifier
                    .padding(end = 16.dp)
                    .size(40.dp)
                    .clip(CircleShape)
                    .background(MaterialTheme.colorScheme.primaryContainer)
                    .clickable { onProfileClick() },
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
                            fontSize = 16.sp
                        )
                    } else {
                        Icon(
                            imageVector = Icons.Default.Person,
                            contentDescription = "Profile",
                            tint = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
        },
        colors = TopAppBarDefaults.topAppBarColors(
            containerColor = MaterialTheme.colorScheme.background
        )
    )
}

@Composable
internal fun KalaTabRow(
    tabs: List<String>,
    selectedIndex: Int,
    onTabSelected: (Int) -> Unit
) {
    TabRow(
        selectedTabIndex = selectedIndex,
        containerColor = MaterialTheme.colorScheme.background,
        contentColor = SelectedPurple,
        indicator = { tabPositions ->
            TabRowDefaults.SecondaryIndicator(
                Modifier.tabIndicatorOffset(tabPositions[selectedIndex]),
                color = SelectedPurple
            )
        },
        divider = {}
    ) {
        tabs.forEachIndexed { index, title ->
            Tab(
                selected = selectedIndex == index,
                onClick = { onTabSelected(index) },
                text = {
                    Text(
                        text = title,
                        fontSize = 11.sp,
                        fontWeight = if (selectedIndex == index) FontWeight.Bold else FontWeight.Normal,
                        color = if (selectedIndex == index) SelectedPurple else SubtitleGray,
                        maxLines = 1,
                        overflow = TextOverflow.Visible,
                        softWrap = false
                    )
                }
            )
        }
    }
}

@Composable
internal fun PostCard(
    post: ArtistPost,
    saved: Boolean,
    onSaveClick: () -> Unit,
    onLikeClick: () -> Unit,
    onDeleteClick: () -> Unit,
    onCommentClick: () -> Unit
) {
    var showMenu by remember { mutableStateOf(false) }
    var fullScreenImageIndex by remember { mutableStateOf<Int?>(null) }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 8.dp),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = SurfaceWhite
        ),
        elevation = CardDefaults.cardElevation(
            defaultElevation = 1.dp
        ),
        border = BorderStroke(1.dp, DividerGray.copy(alpha = 0.4f))
    ) {
        Column(
            modifier = Modifier.fillMaxWidth()
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                ProfileAvatar(
                    initials = post.artistName.toInitials(),
                    imageUrl = post.avatarUrl,
                    avatarBytes = null
                )

                Spacer(Modifier.width(10.dp))

                Column(
                    modifier = Modifier.weight(1f)
                ) {
                    Text(
                        text = post.artistName,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = TextPrimary
                    )

                    Text(
                        text = "${post.craft} • ${post.location} • ${post.timeAgo}",
                        fontSize = 12.sp,
                        color = TextSecondary
                    )
                }

                if (post.isMine) {
                    Box {
                        IconButton(onClick = { showMenu = true }) {
                            Icon(Icons.Default.MoreVert, contentDescription = "More options", tint = TextSecondary)
                        }
                        DropdownMenu(
                            expanded = showMenu,
                            onDismissRequest = { showMenu = false }
                        ) {
                            DropdownMenuItem(
                                text = { Text("Delete Post", color = Color.Red) },
                                onClick = {
                                    showMenu = false
                                    onDeleteClick()
                                },
                                leadingIcon = {
                                    Icon(Icons.Default.Delete, contentDescription = null, tint = Color.Red)
                                }
                            )
                        }
                    }
                }
            }

            Text(
                text = post.caption,
                fontSize = 14.sp,
                color = TextPrimary,
                maxLines = 4,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.padding(horizontal = 14.dp)
            )

            if (post.images.isNotEmpty()) {
                Spacer(Modifier.height(10.dp))
                ImageCarousel(
                    images = post.images,
                    onImageClick = { index -> fullScreenImageIndex = index }
                )
            }

            Spacer(Modifier.height(10.dp))

            Row(
                modifier = Modifier.padding(horizontal = 14.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                IconButton(onClick = onLikeClick) {
                    Icon(
                        imageVector = if (post.isLiked) Icons.Filled.Favorite else Icons.Outlined.FavoriteBorder,
                        contentDescription = "Like",
                        tint = if (post.isLiked) Color.Red else MaterialTheme.colorScheme.onSurface
                    )
                }
                Text(text = "${post.likes}", fontSize = 13.sp)
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

                Text("•", color = TextSecondary)

                Text(
                    text = "${post.comments} comments",
                    fontSize = 12.sp,
                    color = TextSecondary,
                    modifier = Modifier.clickable { onCommentClick() }
                )
            }

            Spacer(Modifier.height(8.dp))

                IconButton(onClick = onSaveClick) {
                    Icon(
                        imageVector = if (saved) Icons.Filled.Bookmark else Icons.Outlined.BookmarkBorder,
                        contentDescription = "Save",
                        tint = if (saved) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface
                    )
                }
            }
            HorizontalDivider(
                color = DividerGray,
                thickness = 1.dp
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 8.dp, vertical = 4.dp),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                FeedPostActionButton(
                    icon = if (post.isLiked) Icons.Filled.Favorite else Icons.Outlined.FavoriteBorder,
                    label = "Like",
                    tint = if (post.isLiked) Color.Red else TextSecondary,
                    onClick = onLikeClick
                )

                FeedPostActionButton(
                    icon = Icons.Default.Star,
                    label = "Comment",
                    onClick = onCommentClick
                )

                FeedPostActionButton(
                    icon = Icons.Default.Share,
                    label = "Share",
                    onClick = { /* Share logic */ }
                )
                var expanded by remember { mutableStateOf(false) }
                Text(
                    buildAnnotatedString {
                        withStyle(
                            style = SpanStyle(fontWeight = FontWeight.Bold)
                        ) {
                            append(post.artistName)
                        }

                        append(" · ")
                        append(post.caption)
                    },
                    fontSize = 13.sp,
                    maxLines = if (expanded) Int.MAX_VALUE else 2,
                    overflow = TextOverflow.Ellipsis
                )
            }
        }


    if (fullScreenImageIndex != null) {
        FullScreenImageViewer(
            images = post.images,
            initialPage = fullScreenImageIndex!!,
            onDismiss = { fullScreenImageIndex = null }
        )
    }
}

@Composable
private fun ImageCarousel(
    images: List<String>,
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
fun FeedPostActionButton(
    icon: ImageVector,
    label: String,
    tint: Color = TextSecondary,
    onClick: () -> Unit = {}
) {
    Row(
        modifier = Modifier
            .clip(RoundedCornerShape(8.dp))
            .clickable { onClick() }
            .padding(horizontal = 12.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(6.dp)
    ) {
        Icon(
            imageVector = icon,
            contentDescription = label,
            tint = tint,
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
