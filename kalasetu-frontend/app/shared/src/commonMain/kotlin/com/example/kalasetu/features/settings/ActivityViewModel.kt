package com.example.kalasetu.features.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.kalasetu.features.application.ApplicationRepository
import com.example.kalasetu.features.event.Event
import com.example.kalasetu.features.feed.ArtistPost
import com.example.kalasetu.repository.EventRepository
import com.example.kalasetu.repository.FeedRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import kotlinx.datetime.LocalDate

class ActivityViewModel : ViewModel() {

    private val postRepository = FeedRepository()
    private val eventRepository = EventRepository()

    private val _myPosts = MutableStateFlow<List<ArtistPost>>(emptyList())
    val myPosts: StateFlow<List<ArtistPost>> = _myPosts

    private val _myEvents = MutableStateFlow<List<Event>>(emptyList())
    val myEvents: StateFlow<List<Event>> = _myEvents

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading

    fun load(userId: String) {
        if (_isLoading.value) return
        viewModelScope.launch {
            _isLoading.value = true

            try {
                val postsResponse = postRepository.getPostsByUser(userId)
                val appPosts = postsResponse.data?.postsByUser.orEmpty().mapNotNull { post ->
                    try {
                        ArtistPost(
                            id = post.id.toIntOrNull() ?: 0,
                            userId = post.userId.toIntOrNull() ?: 0,
                            artistName = post.userName,
                            craft = post.categoryName ?: "",
                            location = "",
                            timeAgo = post.createdAt,
                            avatarUrl = "",
                            images = post.media.sortedBy { it.sortOrder }.map { it.url },
                            caption = post.content,
                            likes = post.likeCount,
                            comments = post.commentCount,
                            isLiked = post.isLikedByMe
                        )
                    } catch (e: Exception) {
                        println("ACTIVITY POST MAPPING ERROR: ${e.message}")
                        null
                    }
                }
                _myPosts.value = appPosts
            } catch (e: Exception) {
                println("ACTIVITY POSTS ERROR: ${e.message}")
            }

            try {
                ApplicationRepository.fetchMyApplications()
            } catch (e: Exception) {
                println("ACTIVITY APPLICATIONS ERROR: ${e.message}")
            }

            try {
                val eventsResponse = eventRepository.getUserEvents()
                val appEvents = eventsResponse.data?.userEvents.orEmpty().mapNotNull { event ->
                    try {
                        Event(
                            id = event.id,
                            title = event.name,
                            description = "",
                            location = "",
                            startDate = LocalDate.parse(event.startDate),
                            endDate = null,
                            organizerName = event.hostName ?: ""
                        )
                    } catch (e: Exception) {
                        println("ACTIVITY EVENT MAPPING ERROR: ${e.message}")
                        null
                    }
                }
                _myEvents.value = appEvents
            } catch (e: Exception) {
                println("ACTIVITY EVENTS ERROR: ${e.message}")
            } finally {
                _isLoading.value = false
            }
        }
    }
}