package com.example.kalasetu.repository

import com.apollographql.apollo.api.DefaultUpload
import com.apollographql.apollo.api.Optional
import com.example.kalasetu.CreatePostMutation
import com.example.kalasetu.data.ApiClient
import com.example.kalasetu.type.CreatePostInput

class PostRepository {
    private val apolloClient = ApiClient.apolloClient

    suspend fun createPost(
        content: String,
        images: List<ByteArray>
    ): Boolean {
        if (images.isNotEmpty()) {
            val upload = images.map { bytes ->
                DefaultUpload.Builder()
                    .content(bytes)
                    .contentType("image/jpeg")
                    .fileName("post_${bytes.hashCode()}.jpg")
                    .build()
            }

            val withMedia = try {
                apolloClient
                    .mutation(
                        CreatePostMutation(
                            CreatePostInput(
                                content = content,
                                media = Optional.presentIfNotNull(upload)
                            )
                        )
                    )
                    .execute()
            } catch (e: Exception) {
                println("========== CREATE POST WITH MEDIA ERROR ==========")
                println(e.message)
                null
            }

            if (withMedia?.data != null && withMedia.errors.isNullOrEmpty()) {
                println("========== POST CREATED WITH MEDIA ==========")
                return true
            }

            println("========== POST WITH MEDIA FAILED, RETRYING CONTENT ONLY ==========")
            println("errors = ${withMedia?.errors}")
        }

        return try {
            val contentOnly = apolloClient
                .mutation(
                    CreatePostMutation(
                        CreatePostInput(
                            content = content
                        )
                    )
                )
                .execute()

            val success = contentOnly.data != null && contentOnly.errors.isNullOrEmpty()

            println("========== CREATE POST (CONTENT ONLY) ==========")
            println("success = $success")
            println("errors = ${contentOnly.errors}")

            success
        } catch (e: Exception) {
            println("========== CREATE POST ERROR ==========")
            println(e.message)
            e.printStackTrace()
            false
        }
    }
}