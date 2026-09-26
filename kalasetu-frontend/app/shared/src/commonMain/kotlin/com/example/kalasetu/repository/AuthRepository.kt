package com.example.kalasetu.repository

import com.example.kalasetu.features.auth.AuthStore
import io.ktor.client.HttpClient
import io.ktor.client.request.header
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.contentType
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.intOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import com.example.kalasetu.BuildKonfig
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.serialization.kotlinx.json.json

class AuthRepository {

    private val client = HttpClient(){
        install(ContentNegotiation) {
            json(
                Json {
                    ignoreUnknownKeys = true
                    isLenient = true
                }
            )
        }
    }

    private val baseUrl = BuildKonfig.API_BASE_URL

    suspend fun login(
        email: String,
        password: String
    ): Result<Unit> {

        return try {

            val response = client.post("$baseUrl/api/v1/auth/login") {

                contentType(ContentType.Application.Json)

                setBody(
                    """
                    {
                        "email": "${email.replace("\"", "\\\"")}",
                        "password": "${password.replace("\"", "\\\"")}"
                    }
                    """.trimIndent()
                )
            }

            val responseText = response.bodyAsText()

            println("========== LOGIN RESPONSE ==========")
            println("status = ${response.status}")
            println("body = $responseText")

            if (response.status.value !in 200..299) {

                return Result.failure(
                    Exception("Login failed: ${response.status}")
                )
            }

            val json = Json {
                ignoreUnknownKeys = true
            }

            val body = json
                .parseToJsonElement(responseText)
                .jsonObject

            val accessToken =
                body["access_token"]
                    ?.jsonPrimitive
                    ?.content

            val refreshToken =
                body["refresh_token"]
                    ?.jsonPrimitive
                    ?.content

            val user =
                body["user"]
                    ?.jsonObject

            if (accessToken.isNullOrBlank()) {

                return Result.failure(
                    Exception("No access token received")
                )
            }

            AuthStore.saveSession(
                accessToken = accessToken,
                refreshToken = refreshToken,
                userId = user?.get("id")?.jsonPrimitive?.intOrNull,
                userEmail = user?.get("email")?.jsonPrimitive?.content,
                userName = user?.get("name")?.jsonPrimitive?.content
            )

            println("========== LOGIN SUCCESS ==========")
            println("userId = ${AuthStore.userId}")
            println("userEmail = ${AuthStore.userEmail}")
            println("token received = true")

            Result.success(Unit)

        } catch (e: Exception) {

            println("========== LOGIN ERROR ==========")
            println(e.message)
            e.printStackTrace()

            Result.failure(e)
        }
    }

    suspend fun changePassword(
        oldPassword: String,
        newPassword: String
    ): Result<Unit> {

        return try {

            val token = AuthStore.accessToken
                ?: return Result.failure(Exception("Not logged in"))

            val response = client.post("$baseUrl/api/v1/auth/change-password") {

                contentType(ContentType.Application.Json)

                header("Authorization", "Bearer $token")

                setBody(
                    """
                    {
                        "current_password": "${oldPassword.replace("\"", "\\\"")}",
                        "new_password": "${newPassword.replace("\"", "\\\"")}"
                    }
                    """.trimIndent()
                )
            }

            val responseText = response.bodyAsText()

            if (response.status.value !in 200..299) {

                return Result.failure(
                    Exception(responseText.ifBlank { "Password change failed: ${response.status}" })
                )
            }

            Result.success(Unit)

        } catch (e: Exception) {

            println("========== CHANGE PASSWORD ERROR ==========")
            println(e.message)
            e.printStackTrace()

            Result.failure(e)
        }
    }

    suspend fun register(
        name: String,
        email: String,
        password: String
    ): Result<Unit> {

        return try {

            val response = client.post("$baseUrl/api/v1/auth/register") {

                contentType(ContentType.Application.Json)

                setBody(
                    """
                    {
                        "name": "${name.replace("\"", "\\\"")}",
                        "email": "${email.replace("\"", "\\\"")}",
                        "password": "${password.replace("\"", "\\\"")}"
                    }
                    """.trimIndent()
                )
            }

            val responseText = response.bodyAsText()

            println("========== REGISTER RESPONSE ==========")
            println("status = ${response.status}")
            println("body = $responseText")

            if (response.status.value !in 200..299) {

                return Result.failure(
                    Exception("Registration failed: ${response.status}")
                )
            }

            println("========== REGISTER SUCCESS ==========")

            Result.success(Unit)

        } catch (e: Exception) {

            println("========== REGISTER ERROR ==========")
            println(e.message)
            e.printStackTrace()

            Result.failure(e)
        }
    }
    suspend fun sendOtp(email: String): Result<Unit> {
        return try {
            val response = client.post("$baseUrl/api/v1/auth/send-otp") {
                contentType(ContentType.Application.Json)

                setBody(
                    buildJsonObject {
                        put("email", email)
                        put("purpose", "signup")
                    }
                )
            }

            if (response.status.isSuccess()) {
                Result.success(Unit)
            } else {
                val body = response.bodyAsText()
                Result.failure(
                    Exception(body.ifBlank { "Failed to send OTP" })
                )
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun verifyOtp(
        email: String,
        otp: String
    ): Result<Unit> {
        return try {
            val response = client.post("$baseUrl/api/v1/auth/verify-otp") {
                contentType(ContentType.Application.Json)

                setBody(
                    buildJsonObject {
                        put("email", email)
                        put("otp", otp)
                        put("purpose", "signup")
                    }
                )
            }

            if (response.status.isSuccess()) {
                Result.success(Unit)
            } else {
                val body = response.bodyAsText()
                Result.failure(
                    Exception(body.ifBlank { "Invalid OTP" })
                )
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun resendOtp(email: String): Result<Unit> {
        return sendOtp(email)
    }
}