package com.example.kalasetu.features.opportunity

import kotlinx.datetime.LocalDate

enum class OpportunityStatus(val label: String) {
    ACTIVE("ACTIVE"),
    DRAFT("DRAFT")
}

data class Opportunity(
    val id: String,
    val eventId: String,
    val title: String,
    val description: String = "",
    val categories: List<String> = emptyList(),
    val location: String = "",
    val startDate: LocalDate? = null,
    val endDate: LocalDate? = null,
    val totalPositions: Int = 0,
    val openSlots: Int = 0,
    val applicationsCount: Int = 0,
    val status: OpportunityStatus = OpportunityStatus.DRAFT
)
