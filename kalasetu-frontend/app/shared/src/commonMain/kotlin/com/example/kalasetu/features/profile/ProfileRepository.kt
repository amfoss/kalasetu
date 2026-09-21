package com.example.kalasetu.features.profile

import com.example.kalasetu.GetProfileQuery
import com.example.kalasetu.data.ApiClient

open class ProfileRepository(
    private val initialProfile: Profile? = null
) : ProfileRepositoryContract {

    override suspend fun fetchProfile(userId: String): Profile {
        val fallback = buildFallback(userId)

        return try {
            val response = ApiClient.apolloClient
                .query(GetProfileQuery(userId = userId))
                .execute()

            println("========== GET PROFILE RESPONSE ==========")
            println("data = ${response.data}")
            println("errors = ${response.errors}")
            println("exception = ${response.exception}")

            val profile = response.data?.profile ?: return fallback

            fallback.copy(
                id = profile.id.ifBlank { fallback.id },
                name = profile.name.ifBlank { fallback.name },
                location = profile.location.ifBlank { fallback.location },
                bio = profile.bio.ifBlank { fallback.bio },
                email = profile.email.ifBlank { fallback.email },
                avatarUrl = profile.avatarUrl ?: fallback.avatarUrl,
                avatarBytes = fallback.avatarBytes,
                followers = profile.followers,
                following = profile.following,
                artworksCount = profile.artworksCount,
                totalLikes = profile.totalLikes,
                skills = profile.skills.ifEmpty { fallback.skills },
                artworksImages = profile.artworksImages.ifEmpty { fallback.artworksImages },
                achievements = profile.achievements.mapNotNull { gql ->
                    gql?.let {
                        Achievement(
                            title = it.title,
                            description = it.description,
                            iconType = when (it.iconType) {
                                com.example.kalasetu.type.AchievementIcon.TOP_CREATOR ->
                                    AchievementIcon.TOP_CREATOR

                                com.example.kalasetu.type.AchievementIcon.FOLLOWERS ->
                                    AchievementIcon.FOLLOWERS

                                com.example.kalasetu.type.AchievementIcon.FEATURED ->
                                    AchievementIcon.FEATURED

                                com.example.kalasetu.type.AchievementIcon.UNKNOWN__ ->
                                    AchievementIcon.TOP_CREATOR
                            }
                        )
                    }
                }.ifEmpty { fallback.achievements },
                recentPosts = profile.recentPosts.map { gql ->
                    ProfilePost(
                        id = gql.id,
                        content = gql.content,
                        mediaType = gql.mediaType,
                        mediaUri = gql.mediaUri,
                        likeCount = gql.likeCount,
                        commentCount = gql.commentCount,
                        createdAt = gql.createdAt
                    )
                }
            )

        } catch (e: Exception) {
            println("========== PROFILE FETCH ERROR ==========")
            println(e.message)
            e.printStackTrace()
            fallback
        }
    }

    private fun buildFallback(userId: String): Profile =
        Profile(
            id = userId,
            name = initialProfile?.name.orEmpty(),
            username = initialProfile?.username.orEmpty(),
            location = initialProfile?.location.orEmpty(),
            bio = initialProfile?.bio.orEmpty(),
            email = initialProfile?.email.orEmpty(),
            avatarUrl = initialProfile?.avatarUrl,
            avatarBytes = initialProfile?.avatarBytes,
        )
}