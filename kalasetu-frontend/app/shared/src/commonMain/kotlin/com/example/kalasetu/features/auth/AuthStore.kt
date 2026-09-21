package com.example.kalasetu.features.auth

object AuthStore {

    var accessToken: String? = null
    var refreshToken: String? = null

    var userId: Int? = null
    var userEmail: String? = null
    var userName: String? = null

    fun clear() {
        accessToken = null
        refreshToken = null
        userId = null
        userEmail = null
        userName = null
    }
}