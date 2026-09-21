package com.example.kalasetu.navigation

sealed class Screen {
    // Auth Screens
    data object AuthSignup : Screen()
    data object AuthOtp : Screen()
    data object AuthLogin : Screen()

    // Common Onboarding Screens (Welcome, BasicInfo, Location, and Done Screens)
    data object OnboardingWelcome : Screen()
    data object OnboardingBasicInfo : Screen()
    data object OnboardingLocation  : Screen()
    data object OnboardingDone      : Screen()

    data object ArtistExperience    : Screen()
    data object OrganizerType       : Screen()
    data object OrganizerIntent     : Screen()
    data object AudienceInterests   : Screen()

    data class Profile(val userId: String) : Screen()
    data class EditProfile(val userId: String) : Screen()
    // --- Artist Flow ---
    data class ArtistHome(val userId: String) : Screen()
    data class EventDetails(val eventId: String) : Screen()
    data class ApplicationForm(
        val eventId: String,
        val opportunityId: String = "",
        val opportunityTitle: String = ""
    ) : Screen()

    data object Feed : Screen()
    data object Store : Screen()

    data object UploadPost : Screen()
    data class PostPreview(
        val description: String,
        val imageBytes: List<ByteArray>,
        val userName: String,
        val userAvatarUrl: String?,
        val userAvatarBytes: ByteArray?
    ) : Screen()
    // --- Organizer Flow ---
    data object SelectArtistCategories : Screen()
    data class OrganizerHome(val userId: String) : Screen()
    data object CreateEvent : Screen()
    data class OrganizerEventList(val userId: String) : Screen()
    data object TimelineAndLocation : Screen()
    data object ReviewEvent : Screen()

    // Applications
    data class ApplicationStatus(val applicationId: String) : Screen()
    data class MyApplications(val userId: String) : Screen()

    // Organizer views all applications for a specific event
    data class EventApplications(val eventId: String) : Screen()

    // Organizer previews one application (with Accept/Reject)
    data class ApplicationPreview(val applicationId: String) : Screen()

    data class CreateOpportunity(val eventId: String) : Screen()
    data class EditOpportunity(val opportunityId: String, val eventId: String) : Screen()
}