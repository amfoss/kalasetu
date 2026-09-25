package com.example.kalasetu.features.settings

import com.example.kalasetu.features.feed.ArtistPost
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

object SavedStore {
    private val _saved = MutableStateFlow<List<ArtistPost>>(emptyList())
    val saved: StateFlow<List<ArtistPost>> = _saved.asStateFlow()

    fun isSaved(postId: Int): Boolean = _saved.value.any { it.id == postId }

    fun toggle(post: ArtistPost) {
        _saved.value = if (isSaved(post.id)) {
            _saved.value.filterNot { it.id == post.id }
        } else {
            _saved.value + post
        }
    }

    fun remove(postId: Int) {
        _saved.value = _saved.value.filterNot { it.id == postId }
    }

    fun clear() {
        _saved.value = emptyList()
    }
}