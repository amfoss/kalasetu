package com.example.kalasetu.repository

import com.example.kalasetu.OnboardUserMutation
import com.example.kalasetu.data.ApiClient
import com.example.kalasetu.type.OnboardingInput
import com.apollographql.apollo.api.Optional
class OnboardingRepository {

    private val apolloClient = ApiClient.apolloClient

    suspend fun onboardUser(
        name: String,
        role: String,
        location: String,
        labels: List<String>,
        bio: String,
        profilePicture: String
    ) = apolloClient
        .mutation(
            OnboardUserMutation(
                input = OnboardingInput(
                    name = name,
                    role = role,
                    location = location,
                    labels = Optional.present(labels),
                    bio = Optional.Present(bio),
                    profilePicture = Optional.Present(profilePicture)
                )
            )
        )
        .execute()
        .also { response ->
            println("========== ONBOARDING RESPONSE ==========")
            println("data = ${response.data}")
            println("errors = ${response.errors}")
            println("exception = ${response.exception}")
        }
}