package com.example.kalasetu.features.event

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

object EventStore {
    private val _events = MutableStateFlow<List<Event>>(emptyList())
    val events: StateFlow<List<Event>> = _events.asStateFlow()

    fun addEvent(event: Event) {
        _events.value = _events.value + event
    }

    fun clear() {
        _events.value = emptyList()
    }
}