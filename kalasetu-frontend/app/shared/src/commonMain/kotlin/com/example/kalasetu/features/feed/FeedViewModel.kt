package com.example.kalasetu.features.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.kalasetu.repository.FeedRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlin.time.Clock

class FeedViewModel : ViewModel() {

    private val repository = FeedRepository()

    private val _posts = MutableStateFlow<List<ArtistPost>>(emptyList())
    val posts: StateFlow<List<ArtistPost>> = _posts

    private val _comments = MutableStateFlow<List<Comment>>(emptyList())
    val comments: StateFlow<List<Comment>> = _comments

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading

    fun loadPosts() {
        viewModelScope.launch {
            _isLoading.value = true

            try {
                val response = repository.getPosts()

                if (!response.errors.isNullOrEmpty()) {
                    println("FEED FAILED: ${response.errors}")
                    return@launch
                }

                val backendPosts = response.data?.posts.orEmpty()
                val currentUserId = com.example.kalasetu.features.auth.AuthStore.userId ?: 1

                _posts.value = backendPosts.map { post ->
                    val pUserId = post.userId.toIntOrNull() ?: 0
                    ArtistPost(
                        id = post.id.toIntOrNull() ?: 0,
                        userId = pUserId,
                        artistName = post.userName,
                        craft = post.categoryName ?: "",
                        location = "",
                        timeAgo = post.createdAt,
                        avatarUrl = "",
                        images = post.media
                            .sortedBy { it.sortOrder }
                            .map { it.url },
                        caption = post.content,
                        likes = post.likeCount,
                        comments = post.commentCount,
                        isLiked = post.isLikedByMe,
                        isMine = (pUserId == currentUserId || pUserId == 0 || (currentUserId == 1 && pUserId == 1))
                    )
                }

            } catch (e: Exception) {
                println("FEED ERROR: ${e.message}")
                e.printStackTrace()
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun toggleLike(postId: Int) {
        val currentPost = _posts.value.firstOrNull { it.id == postId } ?: return
        val currentlyLiked = currentPost.isLiked
        val newLikes = if (currentlyLiked) (currentPost.likes - 1).coerceAtLeast(0) else currentPost.likes + 1

        _posts.update { list ->
            list.map { if (it.id == postId) it.copy(isLiked = !currentlyLiked, likes = newLikes) else it }
        }

        viewModelScope.launch {
            try {
                if (currentlyLiked) {
                    repository.unLikePost(postId.toString())
                } else {
                    repository.likePost(postId.toString())
                }
            } catch (e: Exception) {
                println("LIKE ERROR: ${e.message}")
            }
        }
    }

    fun deletePost(postId: Int) {
        _posts.update { list -> list.filterNot { it.id == postId } }
        viewModelScope.launch {
            try {
                repository.deletePost(postId.toString())
            } catch (e: Exception) {
                println("DELETE POST ERROR: ${e.message}")
            }
        }
    }

    fun loadComments(postId: Int) {
        viewModelScope.launch {
            try {
                val response = repository.getComments(postId.toString())
                val fetched = response.data?.commentsOfPost.orEmpty().map { c ->
                    Comment(
                        id = c.id.toIntOrNull() ?: 0,
                        userName = c.userName,
                        avatarUrl = "",
                        content = c.content,
                        timeAgo = c.createdAt,
                        likes = 0,
                        isLiked = false
                    )
                }
                _comments.value = fetched
            } catch (e: Exception) {
                println("LOAD COMMENTS ERROR: ${e.message}")
            }
        }
    }

    fun addComment(postId: Int, text: String) {
        if (text.isBlank()) return
        val newComment = Comment(
            id = (Clock.System.now().toEpochMilliseconds() % 100000).toInt(),
            userName = "You",
            avatarUrl = "",
            content = text,
            timeAgo = "Just now",
            likes = 0,
            isLiked = false
        )
        _comments.update { it + newComment }
        _posts.update { list ->
            list.map { if (it.id == postId) it.copy(comments = it.comments + 1) else it }
        }
        viewModelScope.launch {
            try {
                repository.addComment(postId.toString(), text)
            } catch (e: Exception) {
                println("ADD COMMENT ERROR: ${e.message}")
            }
        }
    }
}