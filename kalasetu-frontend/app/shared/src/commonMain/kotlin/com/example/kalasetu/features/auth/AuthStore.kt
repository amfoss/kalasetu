package com.example.kalasetu.features.auth

import com.russhwolf.settings.Settings

object AuthStore {

    private const val ACCESS_TOKEN_KEY = "auth_access_token"
    private const val REFRESH_TOKEN_KEY = "auth_refresh_token"
    private const val USER_ID_KEY = "auth_user_id"
    private const val USER_EMAIL_KEY = "auth_user_email"
    private const val USER_NAME_KEY = "auth_user_name"

    private var settings: Settings? = null

    var accessToken: String? = null
        private set

    var refreshToken: String? = null
        private set

    var userId: Int? = null
        private set

    var userEmail: String? = null
        private set

    var userName: String? = null
        private set

    fun initialize(settings: Settings) {
        this.settings = settings

        accessToken = settings.getStringOrNull(ACCESS_TOKEN_KEY)
        refreshToken = settings.getStringOrNull(REFRESH_TOKEN_KEY)
        userId = settings.getIntOrNull(USER_ID_KEY)
        userEmail = settings.getStringOrNull(USER_EMAIL_KEY)
        userName = settings.getStringOrNull(USER_NAME_KEY)
    }

    fun saveSession(
        accessToken: String,
        refreshToken: String?,
        userId: Int?,
        userEmail: String?,
        userName: String?
    ) {
        this.accessToken = accessToken
        this.refreshToken = refreshToken
        this.userId = userId
        this.userEmail = userEmail
        this.userName = userName

        val storage = settings
            ?: error("AuthStore.initialize() must be called before saveSession()")

        storage.putString(ACCESS_TOKEN_KEY, accessToken)

        if (refreshToken != null) {
            storage.putString(REFRESH_TOKEN_KEY, refreshToken)
        } else {
            storage.remove(REFRESH_TOKEN_KEY)
        }

        if (userId != null) {
            storage.putInt(USER_ID_KEY, userId)
        } else {
            storage.remove(USER_ID_KEY)
        }

        if (userEmail != null) {
            storage.putString(USER_EMAIL_KEY, userEmail)
        } else {
            storage.remove(USER_EMAIL_KEY)
        }

        if (userName != null) {
            storage.putString(USER_NAME_KEY, userName)
        } else {
            storage.remove(USER_NAME_KEY)
        }
    }

    fun clear() {
        accessToken = null
        refreshToken = null
        userId = null
        userEmail = null
        userName = null

        settings?.let {
            it.remove(ACCESS_TOKEN_KEY)
            it.remove(REFRESH_TOKEN_KEY)
            it.remove(USER_ID_KEY)
            it.remove(USER_EMAIL_KEY)
            it.remove(USER_NAME_KEY)
        }
    }

    fun isLoggedIn(): Boolean {
        return !accessToken.isNullOrBlank()
    }
}