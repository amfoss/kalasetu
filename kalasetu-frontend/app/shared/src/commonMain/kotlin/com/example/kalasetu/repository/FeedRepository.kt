package com.example.kalasetu.repository

import com.example.kalasetu.GetPostsQuery
import com.example.kalasetu.GetPostsByUserQuery
import com.example.kalasetu.DeletePostMutation
import com.example.kalasetu.LikePostMutation
import com.example.kalasetu.UnLikePostMutation
import com.example.kalasetu.AddCommentMutation
import com.example.kalasetu.CommentsOfPostQuery
import com.example.kalasetu.type.CreateCommentInput
import com.example.kalasetu.data.ApiClient
import com.apollographql.apollo.api.Optional

class FeedRepository {

    private val apolloClient = ApiClient.apolloClient

    suspend fun getPosts(
        limit: Int = 20,
        offset: Int = 0
    ) =
        apolloClient
            .query(
                GetPostsQuery(
                    limit = Optional.Present(limit),
                    offset = Optional.Present(offset)
                )
            )
            .execute()

    suspend fun getPostsByUser(
        userId: String,
        limit: Int = 20,
        offset: Int = 0
    ) =
        apolloClient
            .query(
                GetPostsByUserQuery(
                    userId = userId,
                    limit = Optional.Present(limit),
                    offset = Optional.Present(offset)
                )
            )
            .execute()

    suspend fun deletePost(id: String) =
        apolloClient.mutation(DeletePostMutation(id)).execute()

    suspend fun likePost(id: String) =
        apolloClient.mutation(LikePostMutation(id)).execute()

    suspend fun unLikePost(id: String) =
        apolloClient.mutation(UnLikePostMutation(id)).execute()

    suspend fun addComment(postId: String, content: String) =
        apolloClient.mutation(AddCommentMutation(CreateCommentInput(postId = postId, content = content))).execute()

    suspend fun getComments(postId: String) =
        apolloClient.query(CommentsOfPostQuery(postId)).execute()
}