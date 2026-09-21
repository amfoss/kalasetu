package com.example.kalasetu.repository

import com.example.kalasetu.features.auth.AuthStore
import io.ktor.client.HttpClient
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

class AuthRepository {

    private val client = HttpClient()

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

            // Store authentication information
            AuthStore.accessToken = accessToken
            AuthStore.refreshToken = refreshToken

            AuthStore.userId =
                user?.get("id")
                    ?.jsonPrimitive
                    ?.intOrNull

            AuthStore.userEmail =
                user?.get("email")
                    ?.jsonPrimitive
                    ?.content

            AuthStore.userName =
                user?.get("name")
                    ?.jsonPrimitive
                    ?.content

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
}