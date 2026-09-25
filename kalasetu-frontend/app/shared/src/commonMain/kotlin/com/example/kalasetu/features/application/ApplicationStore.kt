package com.example.kalasetu.features.application

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

object ApplicationStore {
    private val _applications = MutableStateFlow<List<Application>>(emptyList())
    val applications: StateFlow<List<Application>> = _applications.asStateFlow()

    fun addApplication(app: Application) {
        val current = _applications.value
        val exists = current.any { it.id == app.id }
        if (exists) {
            _applications.value = current.map { if (it.id == app.id) app else it }
        } else {
            _applications.value = current + app
        }
    }

    fun setApplications(apps: List<Application>) {
        _applications.value = apps
    }

    fun applicationsForEvent(eventId: String): List<Application> =
        _applications.value.filter { it.eventId == eventId }

    fun applicationById(id: String): Application? =
        _applications.value.firstOrNull { it.id == id }

    // ✅ The missing function
    fun updateStatus(id: String, status: ApplicationStatus) {
        _applications.value = _applications.value.map {
            if (it.id == id) it.copy(status = status) else it
        }
    }

    fun clear() {
        _applications.value = emptyList()
    }
}