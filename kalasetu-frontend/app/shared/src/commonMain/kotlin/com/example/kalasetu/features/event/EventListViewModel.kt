package com.example.kalasetu.features.event

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.kalasetu.repository.EventRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import kotlinx.datetime.LocalDate

class EventListViewModel : ViewModel() {

    private val repository = EventRepository()

    private val _events = MutableStateFlow<List<Event>>(emptyList())
    val events: StateFlow<List<Event>> = _events

    fun loadEvents(isOrganizer: Boolean) {

        viewModelScope.launch {

            try {

                println("========== LOADING EVENTS ==========")

                val appEvents: List<Event> = if (isOrganizer) {
                    val response = repository.getUserEvents()

                    if (!response.errors.isNullOrEmpty()) {
                        println("========== GET USER EVENTS FAILED ==========")
                        println("errors = ${response.errors}")
                        return@launch
                    }

                    val backendEvents =
                        response.data?.userEvents.orEmpty()

                    println(
                        "User events received: ${backendEvents.size}"
                    )

                    backendEvents.mapNotNull { backendEvent ->

                        try {

                            Event(
                                id = backendEvent.id,
                                title = backendEvent.name,
                                description = "",
                                location = "",
                                startDate = LocalDate.parse(
                                    backendEvent.startDate
                                ),
                                endDate = null,
                                organizerName =
                                    backendEvent.hostName ?: "",
                                email = "",
                                phone = "",
                                coverImageBytes = null,
                                galleryBytes = emptyList(),
                                categories = emptyList()
                            )

                        } catch (e: Exception) {

                            println(
                                "Failed to parse event " +
                                        "${backendEvent.id}: ${e.message}"
                            )

                            null
                        }
                    }

                } else {

                    val response = repository.getEvents()

                    if (!response.errors.isNullOrEmpty()) {
                        println("========== GET EVENTS FAILED ==========")
                        println("errors = ${response.errors}")
                        return@launch
                    }

                    val backendEvents =
                        response.data?.events.orEmpty()

                    println(
                        "All events received: ${backendEvents.size}"
                    )

                    backendEvents.mapNotNull { backendEvent ->

                        try {

                            Event(
                                id = backendEvent.id,
                                title = backendEvent.name,
                                description = "",
                                location = "",
                                startDate = LocalDate.parse(
                                    backendEvent.startDate
                                ),
                                endDate = null,
                                organizerName =
                                    backendEvent.hostName ?: "",
                                email = "",
                                phone = "",
                                coverImageBytes = null,
                                galleryBytes = emptyList(),
                                categories = emptyList()
                            )

                        } catch (e: Exception) {

                            println(
                                "Failed to parse event " +
                                        "${backendEvent.id}: ${e.message}"
                            )

                            null
                        }
                    }
                }

                _events.value = appEvents

                println("========== EVENTS LOADED ==========")
                println(
                    "Events displayed: ${appEvents.size}"
                )

            } catch (e: Exception) {

                println("========== GET EVENTS ERROR ==========")
                println(e.message)
                e.printStackTrace()
            }
        }
    }

    fun addEvent(event: Event) {
        _events.value = _events.value + event
    }

    fun getEventById(id: String): Event? {
        return _events.value.firstOrNull { it.id == id }
    }
}