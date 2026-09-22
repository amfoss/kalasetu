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
import com.example.kalasetu.features.profile.toInitials
import com.example.kalasetu.theme.KalasetuTheme
import com.example.kalasetu.theme.SelectedPurple
import com.example.kalasetu.theme.SubtitleGray
import com.example.kalasetu.theme.UnselectedBorder
import kotlinx.coroutines.launch

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
    onMenuClick: () -> Unit
) {
    CenterAlignedTopAppBar(
        title = {
            Text(
                text = "KalaSetu",
                fontWeight = FontWeight.Bold,
                fontSize = 22.sp
            )
        },
        navigationIcon = {
            IconButton(onClick = onMenuClick) {
                Icon(Icons.Default.Menu, contentDescription = "Menu")
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
private fun PostCard(
    post: ArtistPost,
    onLikeClick: () -> Unit,
    onDeleteClick: () -> Unit,
    onCommentClick: () -> Unit
) {
    var saved by remember { mutableStateOf(false) }
    var showMenu by remember { mutableStateOf(false) }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 8.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        shape = RoundedCornerShape(12.dp),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp),
        border = BorderStroke(1.dp, UnselectedBorder.copy(alpha = 0.4f))
    ) {
        Column {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(12.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                AsyncImage(
                    model = post.avatarUrl,
                    contentDescription = post.artistName,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .size(40.dp)
                        .clip(CircleShape)
                        .background(post.avatarBackground)
                )
                Spacer(modifier = Modifier.width(10.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(post.artistName, fontWeight = FontWeight.Bold, fontSize = 15.sp)
                    Text(
                        text = "${post.craft} • ${post.location} • ${post.timeAgo}",
                        fontSize = 12.sp,
                        color = SubtitleGray
                    )
                }
                if (post.isMine) {
                    Box {
                        IconButton(onClick = { showMenu = true }) {
                            Icon(Icons.Default.MoreVert, contentDescription = "More options")
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

            ImageCarousel(images = post.images)

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 6.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                IconButton(onClick = onLikeClick) {
                    Icon(
                        imageVector = if (post.isLiked) Icons.Filled.Favorite else Icons.Outlined.FavoriteBorder,
                        contentDescription = "Like",
                        tint = if (post.isLiked) Color.Red else Color.Black
                    )
                }
                Text(text = "${post.likes}", fontSize = 13.sp)

                Spacer(modifier = Modifier.width(16.dp))

                IconButton(onClick = onCommentClick) {
                    Icon(Icons.Outlined.ChatBubbleOutline, contentDescription = "Comments")
                }
                Text(
                    text = "${post.comments}",
                    fontSize = 13.sp,
                    modifier = Modifier.clickable { onCommentClick() }
                )

                Spacer(modifier = Modifier.weight(1f))

                IconButton(onClick = { saved = !saved }) {
                    Icon(
                        imageVector = if (saved) Icons.Filled.Bookmark else Icons.Outlined.BookmarkBorder,
                        contentDescription = "Save"
                    )
                }
            }

            Column(
                modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)
            ) {
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
                if (!expanded) {
                    Text(
                        text = "see more",
                        fontSize = 12.sp,
                        color = SubtitleGray,
                        modifier = Modifier.clickable { expanded = true }
                    )
                }
                Spacer(modifier = Modifier.height(6.dp))
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun ImageCarousel(images: List<String>) {

    val pagerState = rememberPagerState(
        pageCount = { images.size }
    )

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(220.dp)
    ) {

        HorizontalPager(
            state = pagerState,
            modifier = Modifier.fillMaxSize()
        ) { page ->

            AsyncImage(
                model = images[page],
                contentDescription = null,
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize(),
                onLoading = {
                    println("IMAGE LOADING: ${images[page]}")
                },
                onSuccess = {
                    println("IMAGE SUCCESS: ${images[page]}")
                },
                onError = {
                    println("IMAGE ERROR: ${images[page]}")
                    println("IMAGE ERROR DETAILS: ${it.result.throwable}")
                }
            )
        }

        // Dots
        if (images.size > 1) {
            Row(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                images.indices.forEach { index ->
                    Box(
                        modifier = Modifier
                            .size(6.dp)
                            .clip(CircleShape)
                            .background(
                                if (index == pagerState.currentPage)
                                    SelectedPurple
                                else
                                    Color.White.copy(alpha = 0.6f)
                            )
                    )
                }
            }
        }
    }
}
