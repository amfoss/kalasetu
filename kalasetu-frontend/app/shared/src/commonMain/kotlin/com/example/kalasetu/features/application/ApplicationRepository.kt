package com.example.kalasetu.features.application

import com.apollographql.apollo.api.Optional
import com.example.kalasetu.GetApplicationsByEventQuery
import com.example.kalasetu.GetMyApplicationsQuery
import com.example.kalasetu.SubmitApplicationMutation
import com.example.kalasetu.UpdateApplicationStatusMutation
import com.example.kalasetu.data.ApiClient
import com.example.kalasetu.type.CreateApplicationInput
import kotlinx.datetime.LocalDate

object ApplicationRepository {

    private fun parseStatus(statusStr: String): ApplicationStatus {
        return when (statusStr.lowercase()) {
            "accepted" -> ApplicationStatus.ACCEPTED
            "rejected" -> ApplicationStatus.REJECTED
            else -> ApplicationStatus.PENDING
        }
    }

    suspend fun fetchApplicationsForEvent(eventId: String): Result<List<Application>> {
        return try {
            val response = ApiClient.apolloClient.query(GetApplicationsByEventQuery(eventId)).execute()
            val data = response.data?.applicationsByEvent
            if (data != null) {
                val apps = data.map { gApp ->
                    val date = try { gApp.createdAt.take(10).let { LocalDate.parse(it) } } catch (_: Exception) { null }
                    Application(
                        id = gApp.id,
                        eventId = gApp.eventId,
                        applicantName = gApp.applicantName ?: "",
                        email = gApp.applicantEmail ?: "",
                        phone = gApp.applicantPhone ?: "",
                        description = gApp.description ?: "",
                        portfolioFileName = gApp.resumeUrl ?: "",
                        status = parseStatus(gApp.status),
                        submittedAt = date
                    )
                }
                apps.forEach { ApplicationStore.addApplication(it) }
                Result.success(apps)
            } else {
                val err = response.errors?.firstOrNull()?.message ?: "Failed to fetch event applications"
                Result.failure(Exception(err))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun fetchMyApplications(): Result<List<Application>> {
        return try {
            val response = ApiClient.apolloClient.query(GetMyApplicationsQuery()).execute()
            val data = response.data?.myApplications
            if (data != null) {
                val apps = data.map { gApp ->
                    val date = try { gApp.createdAt.take(10).let { LocalDate.parse(it) } } catch (_: Exception) { null }
                    Application(
                        id = gApp.id,
                        eventId = gApp.eventId,
                        applicantName = gApp.applicantName ?: "",
                        email = gApp.applicantEmail ?: "",
                        phone = gApp.applicantPhone ?: "",
                        description = gApp.description ?: "",
                        portfolioFileName = gApp.resumeUrl ?: "",
                        status = parseStatus(gApp.status),
                        submittedAt = date
                    )
                }
                apps.forEach { ApplicationStore.addApplication(it) }
                Result.success(apps)
            } else {
                val err = response.errors?.firstOrNull()?.message ?: "Failed to fetch user applications"
                Result.failure(Exception(err))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun submitApplication(
        eventId: String,
        opportunityId: String? = null,
        applicantName: String,
        email: String,
        phone: String,
        description: String,
        resumeUrl: String
    ): Result<Application> {
        return try {
            val input = CreateApplicationInput(
                eventId = eventId,
                opportunityId = Optional.presentIfNotNull(opportunityId?.ifBlank { null }),
                applicantName = Optional.presentIfNotNull(applicantName.ifBlank { null }),
                applicantEmail = Optional.presentIfNotNull(email.ifBlank { null }),
                applicantPhone = Optional.presentIfNotNull(phone.ifBlank { null }),
                description = Optional.presentIfNotNull(description.ifBlank { null }),
                resumeUrl = Optional.presentIfNotNull(resumeUrl.ifBlank { null })
            )

            val response = ApiClient.apolloClient.mutation(SubmitApplicationMutation(input)).execute()
            val gApp = response.data?.submitApplication
            if (gApp != null && !gApp.id.isNullOrBlank()) {
                val date = try { gApp.createdAt.take(10).let { LocalDate.parse(it) } } catch (_: Exception) { null }
                val app = Application(
                    id = gApp.id,
                    eventId = gApp.eventId,
                    applicantName = gApp.applicantName ?: applicantName,
                    email = gApp.applicantEmail ?: email,
                    phone = gApp.applicantPhone ?: phone,
                    description = gApp.description ?: description,
                    portfolioFileName = gApp.resumeUrl ?: resumeUrl,
                    status = parseStatus(gApp.status),
                    submittedAt = date
                )
                // Don't add to store here, let the UI callback handle it to avoid duplication
                Result.success(app)
            } else {
                val err = response.errors?.firstOrNull()?.message ?: "Failed to submit application (No ID returned)"
                Result.failure(Exception(err))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateStatus(id: String, status: ApplicationStatus): Result<Boolean> {
        return try {
            val statusStr = when (status) {
                ApplicationStatus.ACCEPTED -> "accepted"
                ApplicationStatus.REJECTED -> "rejected"
                else -> "pending"
            }
            val response = ApiClient.apolloClient.mutation(UpdateApplicationStatusMutation(id, statusStr)).execute()
            val success = response.data?.updateApplicationStatus == true
            if (success) {
                ApplicationStore.updateStatus(id, status)
                Result.success(true)
            } else {
                val err = response.errors?.firstOrNull()?.message ?: "Failed to update application status"
                Result.failure(Exception(err))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
