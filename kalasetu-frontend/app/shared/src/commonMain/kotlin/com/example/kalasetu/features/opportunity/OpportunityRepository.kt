package com.example.kalasetu.features.opportunity

import com.apollographql.apollo.api.Optional
import com.example.kalasetu.CreateOpportunityMutation
import com.example.kalasetu.UpdateOpportunityMutation
import com.example.kalasetu.GetOpportunitiesByEventQuery
import com.example.kalasetu.data.ApiClient
import com.example.kalasetu.type.CreateOpportunityInput
import com.example.kalasetu.type.UpdateOpportunityInput
import kotlinx.datetime.LocalDate
import kotlinx.datetime.toLocalDate
import kotlin.time.Clock

object OpportunityRepository {

    suspend fun fetchOpportunitiesForEvent(eventId: String): Result<List<Opportunity>> {
        return try {
            val response = ApiClient.apolloClient.query(GetOpportunitiesByEventQuery(eventId)).execute()
            val data = response.data?.opportunitiesByEvent
            if (data != null) {
                val opps = data.map { gOpp ->
                    val sDate = try { gOpp.startDate?.let { LocalDate.parse(it) } } catch (_: Exception) { null }
                    val eDate = try { gOpp.endDate?.let { LocalDate.parse(it) } } catch (_: Exception) { null }
                    
                    Opportunity(
                        id = gOpp.id,
                        eventId = gOpp.eventId,
                        title = gOpp.title,
                        description = gOpp.description ?: "",
                        categories = gOpp.categories,
                        location = gOpp.location ?: "",
                        startDate = sDate,
                        endDate = eDate,
                        totalPositions = gOpp.totalPositions,
                        openSlots = gOpp.openSlots,
                        applicationsCount = gOpp.applicationsCount,
                        status = if (gOpp.status == "DRAFT") OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
                    )
                }
                opps.forEach { OpportunityStore.addOpportunity(it) }
                Result.success(opps)
            } else {
                val err = response.errors?.firstOrNull()?.message ?: "Failed to fetch opportunities"
                Result.failure(Exception(err))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun createOpportunity(
        eventId: String,
        title: String,
        description: String,
        categories: List<String>,
        location: String,
        startDate: LocalDate?,
        endDate: LocalDate?,
        totalPositions: Int,
        isDraft: Boolean
    ): Result<Opportunity> {
        return try {
            val input = CreateOpportunityInput(
                eventId = eventId,
                title = title,
                description = Optional.presentIfNotNull(description.ifBlank { null }),
                categories = Optional.presentIfNotNull(categories.ifEmpty { null }),
                location = Optional.presentIfNotNull(location.ifBlank { null }),
                startDate = Optional.presentIfNotNull(startDate?.toString()),
                endDate = Optional.presentIfNotNull(endDate?.toString()),
                totalPositions = Optional.presentIfNotNull(totalPositions),
                status = Optional.presentIfNotNull(if (isDraft) "DRAFT" else "ACTIVE")
            )

            val response = ApiClient.apolloClient.mutation(CreateOpportunityMutation(input)).execute()
            val gOpp = response.data?.createOpportunity
            if (gOpp != null) {
                val sDate = try { gOpp.startDate?.let { LocalDate.parse(it) } } catch (_: Exception) { null }
                val eDate = try { gOpp.endDate?.let { LocalDate.parse(it) } } catch (_: Exception) { null }

                val opp = Opportunity(
                    id = gOpp.id,
                    eventId = gOpp.eventId,
                    title = gOpp.title,
                    description = gOpp.description ?: "",
                    categories = gOpp.categories,
                    location = gOpp.location ?: "",
                    startDate = sDate,
                    endDate = eDate,
                    totalPositions = gOpp.totalPositions,
                    openSlots = gOpp.openSlots,
                    applicationsCount = gOpp.applicationsCount,
                    status = if (gOpp.status == "DRAFT") OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
                )
                OpportunityStore.addOpportunity(opp)
                Result.success(opp)
            } else {
                val err = response.errors?.firstOrNull()?.message ?: "Failed to create opportunity"
                println("========== CREATE OPPORTUNITY WARNING: $err ==========")
                val localOpp = Opportunity(
                    id = "opp_${Clock.System.now().toEpochMilliseconds()}",
                    eventId = eventId,
                    title = title,
                    description = description,
                    categories = categories,
                    location = location,
                    startDate = startDate,
                    endDate = endDate,
                    totalPositions = totalPositions,
                    openSlots = totalPositions,
                    status = if (isDraft) OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
                )
                OpportunityStore.addOpportunity(localOpp)
                Result.success(localOpp)
            }
        } catch (e: Exception) {
            println("========== CREATE OPPORTUNITY EXCEPTION: ${e.message} ==========")
            val localOpp = Opportunity(
                id = "opp_${Clock.System.now().toEpochMilliseconds()}",
                eventId = eventId,
                title = title,
                description = description,
                categories = categories,
                location = location,
                startDate = startDate,
                endDate = endDate,
                totalPositions = totalPositions,
                openSlots = totalPositions,
                status = if (isDraft) OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
            )
            OpportunityStore.addOpportunity(localOpp)
            Result.success(localOpp)
        }
    }

    suspend fun updateOpportunity(
        id: String,
        title: String,
        description: String,
        categories: List<String>,
        location: String,
        startDate: LocalDate?,
        endDate: LocalDate?,
        totalPositions: Int,
        isDraft: Boolean
    ): Result<Opportunity> {
        return try {
            val input = UpdateOpportunityInput(
                title = Optional.presentIfNotNull(title.ifBlank { null }),
                description = Optional.presentIfNotNull(description.ifBlank { null }),
                categories = Optional.presentIfNotNull(categories.ifEmpty { null }),
                location = Optional.presentIfNotNull(location.ifBlank { null }),
                startDate = Optional.presentIfNotNull(startDate?.toString()),
                endDate = Optional.presentIfNotNull(endDate?.toString()),
                totalPositions = Optional.presentIfNotNull(totalPositions),
                status = Optional.presentIfNotNull(if (isDraft) "DRAFT" else "ACTIVE")
            )

            val response = ApiClient.apolloClient.mutation(UpdateOpportunityMutation(id, input)).execute()
            val gOpp = response.data?.updateOpportunity
            if (gOpp != null) {
                val sDate = try { gOpp.startDate?.let { LocalDate.parse(it) } } catch (_: Exception) { null }
                val eDate = try { gOpp.endDate?.let { LocalDate.parse(it) } } catch (_: Exception) { null }

                val opp = Opportunity(
                    id = gOpp.id,
                    eventId = gOpp.eventId,
                    title = gOpp.title,
                    description = gOpp.description ?: "",
                    categories = gOpp.categories,
                    location = gOpp.location ?: "",
                    startDate = sDate,
                    endDate = eDate,
                    totalPositions = gOpp.totalPositions,
                    openSlots = gOpp.openSlots,
                    applicationsCount = gOpp.applicationsCount,
                    status = if (gOpp.status == "DRAFT") OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
                )
                OpportunityStore.updateOpportunity(opp)
                Result.success(opp)
            } else {
                val existing = OpportunityStore.getOpportunity(id)
                val updated = existing?.copy(
                    title = title,
                    description = description,
                    categories = categories,
                    location = location,
                    startDate = startDate,
                    endDate = endDate,
                    totalPositions = totalPositions,
                    status = if (isDraft) OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
                ) ?: Opportunity(
                    id = id,
                    eventId = "1",
                    title = title,
                    description = description,
                    categories = categories,
                    location = location,
                    startDate = startDate,
                    endDate = endDate,
                    totalPositions = totalPositions,
                    openSlots = totalPositions,
                    status = if (isDraft) OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
                )
                OpportunityStore.updateOpportunity(updated)
                Result.success(updated)
            }
        } catch (e: Exception) {
            val existing = OpportunityStore.getOpportunity(id)
            val updated = existing?.copy(
                title = title,
                description = description,
                categories = categories,
                location = location,
                startDate = startDate,
                endDate = endDate,
                totalPositions = totalPositions,
                status = if (isDraft) OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
            ) ?: Opportunity(
                id = id,
                eventId = "1",
                title = title,
                description = description,
                categories = categories,
                location = location,
                startDate = startDate,
                endDate = endDate,
                totalPositions = totalPositions,
                openSlots = totalPositions,
                status = if (isDraft) OpportunityStatus.DRAFT else OpportunityStatus.ACTIVE
            )
            OpportunityStore.updateOpportunity(updated)
            Result.success(updated)
        }
    }
}
