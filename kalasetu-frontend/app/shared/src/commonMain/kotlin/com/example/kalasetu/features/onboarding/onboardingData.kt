package com.example.kalasetu.features.onboarding


data class OnboardingData(
    val name: String = "",
    val role: String = "",
    val location: String = "",
    val labels: List<String> = emptyList(),
    val bio: String = "",
    val profilePicture: String = ""
)