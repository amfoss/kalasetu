package com.example.kalasetu.data

import com.apollographql.apollo.ApolloClient
import com.apollographql.apollo.api.ApolloRequest
import com.apollographql.apollo.api.ApolloResponse
import com.apollographql.apollo.api.Operation
import com.apollographql.apollo.interceptor.ApolloInterceptor
import com.apollographql.apollo.interceptor.ApolloInterceptorChain
import com.example.kalasetu.features.auth.AuthStore
import kotlinx.coroutines.flow.Flow
import com.example.kalasetu.BuildKonfig
class AuthorizationInterceptor : ApolloInterceptor {

    override fun <D : Operation.Data> intercept(
        request: ApolloRequest<D>,
        chain: ApolloInterceptorChain
    ): Flow<ApolloResponse<D>> {

        val token = AuthStore.accessToken

        return if (!token.isNullOrBlank()) {

            println("========== APOLLO AUTH ==========")
            println("Access token available = true")
            println("Adding Authorization header")

            val authenticatedRequest = request
                .newBuilder()
                .addHttpHeader(
                    "Authorization",
                    "Bearer $token"
                )
                .build()

            chain.proceed(authenticatedRequest)

        } else {

            println("========== APOLLO AUTH ==========")
            println("Access token available = false")
            println("Sending request without Authorization header")

            chain.proceed(request)
        }
    }
}

object ApiClient {

    val apolloClient: ApolloClient = ApolloClient.Builder()
        .serverUrl("${BuildKonfig.API_BASE_URL}/api/v1/graphql")
        .addInterceptor(AuthorizationInterceptor())
        .build()
}