package com.example.kalasetu.features.profile

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp

@Composable
fun ProfileScreen(
    presenter: ProfilePresenter,
    userId: String,
    posts: List<DraftPost>,
    onEditProfile: () -> Unit = {},
    onShare: () -> Unit = {},
    onBack: () -> Unit = {},
) {
    var uiState by remember { mutableStateOf(ProfileUiState()) }

    val view = remember {
        object : ProfileContract.View {

            override fun showLoading(isLoading: Boolean) {
                uiState = uiState.copy(isLoading = isLoading)
            }

            override fun showProfile(profile: Profile) {
                uiState = uiState.copy(
                    profile = profile,
                    error = null
                )
            }

            override fun showError(message: String) {
                uiState = uiState.copy(error = message)
            }
        }
    }

    LaunchedEffect(userId, presenter) {
        presenter.attachView(view)
        presenter.loadProfile(userId)
    }

    DisposableEffect(Unit) {
        onDispose {
            presenter.detach()
        }
    }

    when {
        uiState.isLoading -> LoadingContent()

        uiState.error != null -> {
            ErrorContent(message = uiState.error!!)
        }

        uiState.profile != null -> {
            val backendPosts = uiState.profile!!.recentPosts.map { post -> profilePostToDraftPost(post) }
            val displayPosts = (posts + backendPosts).distinctBy { it.content }

            ProfileContent(
                profile = uiState.profile!!,
                posts = displayPosts,
                selectedTab = uiState.selectedTab,
                onTabSelected = { tab ->
                    uiState = uiState.copy(selectedTab = tab)
                },
                onEditProfile = onEditProfile,
                onShare = onShare,
                onBack = onBack
            )
        }
    }
}

@Composable
private fun ProfileContent(
    profile: Profile,
    selectedTab: ProfileTab,
    posts: List<DraftPost>,
    onTabSelected: (ProfileTab) -> Unit,
    onEditProfile: () -> Unit,
    onShare: () -> Unit,
    onBack: () -> Unit
) {
    val scrollState = rememberScrollState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(SurfaceWhite)
            .verticalScroll(scrollState)
    ) {
        ProfileHeader(
            profile = profile,
            onEditProfile = onEditProfile,
            onShare = onShare,
            onBack = onBack
        )

        ProfileInfo(profile = profile)

        Spacer(Modifier.height(16.dp))

        StatCardsRow(profile = profile)

        Spacer(Modifier.height(16.dp))

        ProfileTabBar(
            selected = selectedTab,
            onTabSelected = onTabSelected
        )

        Spacer(Modifier.height(12.dp))

        when (selectedTab) {
            ProfileTab.POSTS ->
                PostsTabContent(
                    profile = profile,
                    posts = posts
                )

            ProfileTab.SKILLS ->
                SkillsTabContent(profile.skills)

            ProfileTab.ACHIEVEMENTS ->
                AchievementsTabContent(profile.achievements)
        }

        Spacer(Modifier.height(16.dp))

        ContactSocialCard(profile = profile)

        Spacer(Modifier.height(32.dp))
    }
}

@Composable
private fun LoadingContent() {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        CircularProgressIndicator()
    }
}

@Composable
private fun ErrorContent(message: String) {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = message,
            textAlign = TextAlign.Center
        )
    }
}

@Composable
fun EmptyTabMessage(text: String) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(20.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(text = text)
    }
}