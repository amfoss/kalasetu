package com.example.kalasetu.repository

import com.example.kalasetu.CreateEventMutation
import com.example.kalasetu.data.ApiClient
import com.example.kalasetu.type.CreateEventInput
import com.example.kalasetu.GetEventsQuery
import com.example.kalasetu.GetUserEventsQuery
import com.example.kalasetu.BuildKonfig

private val baseUrl = BuildKonfig.API_BASE_URL
class EventRepository {
    private val apolloClient = ApiClient.apolloClient

    suspend fun createEvent(
        name: String,
        startDate: String,
        duration: String
    ) = apolloClient
        .mutation(
            CreateEventMutation(
                input = CreateEventInput(
                    name = name,
                    startDate = startDate,
                    duration = duration
                )
            )
        )
        .execute()

    suspend fun getEvents() =
        apolloClient
            .query(GetEventsQuery())
            .execute()
            .also { response ->
                println("========== GET EVENTS RESPONSE ==========")
                println("data = ${response.data}")
                println("errors = ${response.errors}")
                println("exception = ${response.exception}")
            }

    suspend fun getUserEvents() =
        apolloClient
            .query(GetUserEventsQuery())
            .execute()
            .also { response ->
                println("========== GET USER EVENTS RESPONSE ==========")
                println("data = ${response.data}")
                println("errors = ${response.errors}")
                println("exception = ${response.exception}")
            }

}