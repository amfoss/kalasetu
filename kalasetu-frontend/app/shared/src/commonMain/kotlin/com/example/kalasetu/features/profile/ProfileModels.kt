package com.example.kalasetu.features.profile

data class Profile(
    val id: String = "",
    val name: String = "",
    val username: String = "",
    val location: String = "",
    val bio: String = "",
    val avatarUrl: String? = null,
    val avatarBytes: ByteArray? = null,
    val followers: Int = 0,
    val following: Int = 0,
    val artworksCount: Int = 0,
    val totalLikes: Int = 0,
    val email: String = "",
    val isVerified: Boolean = false,
    val skills: List<String> = emptyList(),
    val artworksImages: List<String> = emptyList(),
    val achievements: List<Achievement> = emptyList(),
    val recentPosts: List<ProfilePost> = emptyList(),
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if ((other == null) || (this::class != other::class)) return false

        other as Profile

        if (id != other.id) return false
        if (name != other.name) return false
        if (username != other.username) return false
        if (location != other.location) return false
        if (bio != other.bio) return false
        if (avatarUrl != other.avatarUrl) return false
        if (avatarBytes != null) {
            if (other.avatarBytes == null) return false
            if (!avatarBytes.contentEquals(other.avatarBytes)) return false
        } else if (other.avatarBytes != null) return false
        if (followers != other.followers) return false
        if (following != other.following) return false
        if (artworksCount != other.artworksCount) return false
        if (totalLikes != other.totalLikes) return false
        if (email != other.email) return false
        if (isVerified != other.isVerified) return false
        if (skills != other.skills) return false
        if (artworksImages != other.artworksImages) return false
        if (achievements != other.achievements) return false
        return recentPosts == other.recentPosts
    }

    override fun hashCode(): Int {
        var result = id.hashCode()
        result = (31 * result) + name.hashCode()
        result = (31 * result) + username.hashCode()
        result = (31 * result) + location.hashCode()
        result = (31 * result) + bio.hashCode()
        result = (31 * result) + (avatarUrl?.hashCode() ?: 0)
        result = (31 * result) + (avatarBytes?.contentHashCode() ?: 0)
        result = (31 * result) + followers
        result = (31 * result) + following
        result = (31 * result) + artworksCount
        result = (31 * result) + totalLikes
        result = (31 * result) + email.hashCode()
        result = (31 * result) + isVerified.hashCode()
        result = (31 * result) + skills.hashCode()
        result = (31 * result) + artworksImages.hashCode()
        result = (31 * result) + achievements.hashCode()
        result = (31 * result) + recentPosts.hashCode()
        return result
    }
}

data class ProfilePost(
    val id: String = "",
    val content: String = "",
    val mediaType: String? = null,
    val mediaUri: String? = null,
    val likeCount: Int = 0,
    val commentCount: Int = 0,
    val createdAt: String = "",
)

data class Achievement(
    val title: String,
    val description: String,
    val iconType: AchievementIcon = AchievementIcon.TOP_CREATOR,
)

enum class AchievementIcon {
    TOP_CREATOR, FOLLOWERS, FEATURED
}

enum class ProfileTab {
    POSTS, SKILLS, ACHIEVEMENTS
}

data class ProfileUiState(
    val isLoading: Boolean = false,
    val profile: Profile? = null,
    val error: String? = null,
    val selectedTab: ProfileTab = ProfileTab.ACHIEVEMENTS,
)

fun String.toInitials(): String {
    if (isBlank()) return ""
    val parts = trim()
        .split(Regex("[\\s-]+"))
        .filter { it.isNotBlank() }

    if (parts.isEmpty()) return ""

    val firstInitial = parts.first().first().uppercaseChar()
    if (parts.size == 1) return firstInitial.toString()

    val lastInitial = parts.last().first().uppercaseChar()
    return "$firstInitial$lastInitial"
}
fun Int.toDisplayCount(): String = when {
    this >= 1_000_000 -> {
        val m = this / 1_000_000.0
        val s = (m * 10).toInt()
        if ((s % 10) == 0) "${s / 10}M" else "${s / 10}.${s % 10}M"
    }
    this >= 1_000 -> {
        val k = this / 1_000.0
        val s = (k * 10).toInt()
        if ((s % 10) == 0) "${s / 10}K" else "${s / 10}.${s % 10}K"
    }
    else -> this.toString()
}

interface ProfileRepositoryContract {
    suspend fun fetchProfile(userId: String): Profile
}