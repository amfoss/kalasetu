package com.example.kalasetu.features.opportunity

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.datetime.LocalDate

data class CreateOpportunityState(
    val title: String = "",
    val categories: List<String> = emptyList(),
    val location: String = "",
    val startDate: LocalDate? = null,
    val endDate: LocalDate? = null,
    val totalPositions: Int = 1,
    val description: String = "",
    val isSubmitting: Boolean = false,
    val error: String? = null
)

class CreateOpportunityViewModel : ViewModel() {
    private val _state = MutableStateFlow(CreateOpportunityState())
    val state: StateFlow<CreateOpportunityState> = _state.asStateFlow()

    fun onTitleChange(title: String) = _state.update { it.copy(title = title) }
    fun onLocationChange(loc: String) = _state.update { it.copy(location = loc) }
    fun onDescriptionChange(desc: String) = _state.update { it.copy(description = desc) }
    
    fun toggleCategory(cat: String) = _state.update { 
        val newCats = if (it.categories.contains(cat)) it.categories - cat else it.categories + cat
        it.copy(categories = newCats)
    }

    fun onPositionsChange(count: Int) = _state.update { 
        it.copy(totalPositions = count.coerceAtLeast(0))
    }

    fun onDatesChange(start: LocalDate?, end: LocalDate?) = _state.update { 
        it.copy(startDate = start, endDate = end)
    }

    fun submit(eventId: String, isDraft: Boolean, onComplete: () -> Unit = {}) {
        val current = _state.value
        val titleToUse = current.title.ifBlank { "General Opportunity" }

        _state.update { it.copy(isSubmitting = true, error = null) }
        viewModelScope.launch {
            val result = OpportunityRepository.createOpportunity(
                eventId = eventId,
                title = titleToUse,
                description = current.description,
                categories = current.categories,
                location = current.location,
                startDate = current.startDate,
                endDate = current.endDate,
                totalPositions = current.totalPositions,
                isDraft = isDraft
            )
            _state.update { it.copy(isSubmitting = false) }
            onComplete()
        }
    }
}
