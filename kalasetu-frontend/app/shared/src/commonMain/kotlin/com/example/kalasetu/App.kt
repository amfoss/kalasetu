package com.example.kalasetu

import androidx.compose.foundation.layout.*
import androidx.compose.runtime.*
import androidx.compose.material3.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.ui.unit.dp
import com.example.kalasetu.features.auth.*
import com.example.kalasetu.features.feed.*
import com.example.kalasetu.features.application.*
import com.example.kalasetu.features.event.*
import com.example.kalasetu.features.opportunity.CreateOpportunityScreen
import com.example.kalasetu.features.opportunity.EditOpportunityScreen
import com.example.kalasetu.features.marketplace.*
import com.example.kalasetu.features.onboarding.*
import com.example.kalasetu.features.profile.*
import com.example.kalasetu.features.settings.SavedStore
import com.example.kalasetu.features.settings.SettingsScreen
import com.example.kalasetu.features.settings.SettingsStore
import com.example.kalasetu.features.settings.ThemeMode
import com.example.kalasetu.features.settings.rememberSettingsStorage
import com.example.kalasetu.features.opportunity.OpportunityStore
import com.example.kalasetu.navigation.BackHandler
import com.example.kalasetu.navigation.Screen
import com.example.kalasetu.theme.KalasetuTheme
import kotlinx.datetime.LocalDate
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.kalasetu.repository.EventRepository
import kotlinx.coroutines.launch
import com.example.kalasetu.repository.AuthRepository
import com.example.kalasetu.repository.PostRepository
import com.example.kalasetu.repository.OnboardingRepository
import com.example.kalasetu.features.onboarding.OnboardingData
import kotlin.random.Random
import com.russhwolf.settings.Settings

@Composable
fun App() {
    var isRestoringSession by remember { mutableStateOf(true) }
    var screen by remember {mutableStateOf<Screen>(Screen.OnboardingWelcome) }
    var selectedRole by remember { mutableStateOf("") }
    var userName by remember { mutableStateOf("") }
    var userEmail by remember { mutableStateOf("") }
    var userLocation by remember { mutableStateOf("") }
    var signupEmail by remember { mutableStateOf("") }
    var signupOtp by remember { mutableStateOf("") }
    var currentProfile by remember { mutableStateOf<Profile?>(null) }
    var draftEvent by remember { mutableStateOf(EventDraft()) }
    var publishedPosts by remember { mutableStateOf<List<DraftPost>>(emptyList()) }
    val authSettings = remember { Settings() }
    val sharedEventListViewModel: EventListViewModel = viewModel()
    val eventRepository = remember { EventRepository() }
    var onboardingData by remember { mutableStateOf(OnboardingData()) }
    val onboardingRepository = remember { OnboardingRepository() }
    val drawerState = rememberDrawerState(initialValue = DrawerValue.Closed)
    val scope = rememberCoroutineScope()
    val feedViewModel: FeedViewModel = viewModel()
    val marketplaceViewModel: MarketplaceViewModel = viewModel()

    val isAuthScreen = screen is Screen.OnboardingWelcome ||
            screen is Screen.AuthSignupEmail ||
            screen is Screen.AuthOtp ||
            screen is Screen.AuthSignupDetails ||
            screen is Screen.AuthLogin
    val settingsStorage = rememberSettingsStorage()
    remember { SettingsStore.configure(settingsStorage) }
    var themeMode by remember { mutableStateOf(SettingsStore.getThemeMode()) }

    val darkTheme = when (themeMode) {
        ThemeMode.SYSTEM -> isSystemInDarkTheme()
        ThemeMode.LIGHT -> false
        ThemeMode.DARK -> true
    }

    LaunchedEffect(Unit) {
        AuthStore.initialize(authSettings)

        screen = if (AuthStore.isLoggedIn()) {
            Screen.Feed
        } else {
            Screen.OnboardingWelcome
        }

        isRestoringSession = false
    }
    if (isRestoringSession) {
        return
    }
    KalasetuTheme(darkTheme = darkTheme) {
        ModalNavigationDrawer(
            drawerState = drawerState,
            gesturesEnabled = !isAuthScreen,
            drawerContent = {
                SidebarContent(
                    userName = currentProfile?.name ?: userName,
                    userEmail = currentProfile?.email ?: userEmail,
                    userAvatarUrl = currentProfile?.avatarUrl,
                    userAvatarBytes = currentProfile?.avatarBytes,
                    currentRoute = when (screen) {
                        Screen.Feed -> "Dashboard"
                        Screen.Store -> "Store"
                        Screen.Settings -> "Settings"
                        Screen.Marketplace -> "Store"
                        is Screen.ProductOverview -> "Store"
                        is Screen.Profile -> "Profile"
                        is Screen.ArtistHome -> "Events"
                        is Screen.MyApplications -> "Applications"
                        is Screen.OrganizerHome -> "MyEvents"
                        else -> ""
                    },
                    onClose = { scope.launch { drawerState.close() } },
                    onNavigate = { route ->
                        scope.launch { drawerState.close() }
                        when (route) {
                            "Profile" -> screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString())
                            "Settings" -> screen = Screen.Settings
                            "Store" -> screen = Screen.Store
                            "Dashboard" -> screen = Screen.Feed
                            "Events" -> screen = Screen.ArtistHome(userId = "123")
                            "MyEvents" -> screen = Screen.OrganizerHome(userId = "123")
                            "Applications" -> screen = Screen.MyApplications(userId = "123")
                        }
                    }
                )
            }
        ) {
            when (val currentScreen = screen) {
                Screen.Feed -> FeedScreen(
                    viewModel = feedViewModel,
                    userAvatarUrl = currentProfile?.avatarUrl,
                    userAvatarBytes = currentProfile?.avatarBytes,
                    userName = currentProfile?.name ?: userName,
                    onNavigateToProfile = {
                        screen = Screen.Profile(userId = AuthStore.userId?.toString() ?: "")
                    },
                    onNavigateToStore = { screen = Screen.Store },
                    onNavigateToEvents = { screen = Screen.Events },
                    onNavigateToHome = { screen = Screen.Feed },
                    onMenuClick = { scope.launch { drawerState.open() } }
                )

                // ─── Marketplace (Store tab & Audience) ───
                Screen.Store, Screen.Marketplace -> {
                    BackHandler { screen = Screen.Feed }
                    MarketplaceScreen(
                        viewModel = marketplaceViewModel,
                        avatarUrl = currentProfile?.avatarUrl,
                        avatarBytes = currentProfile?.avatarBytes,
                        userName = currentProfile?.name ?: userName,
                        onProductClick = { productId -> screen = Screen.ProductOverview(productId) },
                        onProfileClick = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) },
                        onMenuClick = { scope.launch { drawerState.open() } },
                        onHomeClick = { screen = Screen.Feed },
                        onEventsClick = { screen = Screen.Events },
                        onStoreClick = { screen = Screen.Store },
                    )
                }


                is Screen.ProductOverview -> {
                    BackHandler { screen = Screen.Marketplace }
                    ProductOverviewScreen(
                        productId = currentScreen.productId,
                        viewModel = marketplaceViewModel,
                        avatarUrl = currentProfile?.avatarUrl,
                        avatarBytes = currentProfile?.avatarBytes,
                        userName = currentProfile?.name ?: userName,
                        onBack = { screen = Screen.Marketplace },
                        onProfileClick = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) },
                    )
                }

                Screen.Settings -> SettingsScreen(
                    userName = currentProfile?.name ?: userName,
                    userEmail = currentProfile?.email ?: userEmail,
                    userAvatarUrl = currentProfile?.avatarUrl,
                    userAvatarBytes = currentProfile?.avatarBytes,
                    onNavigateToProfile = {
                        screen = Screen.Profile(userId = AuthStore.userId?.toString() ?: "")
                    },
                    onBack = { screen = Screen.Feed },
                    onMenuClick = { scope.launch { drawerState.open() } },
                    onLoggedOut = {
                        AuthStore.clear()
                        SavedStore.clear()
                        EventStore.clear()
                        ApplicationStore.clear()
                        OpportunityStore.clear()
                        currentProfile = null
                        userName = ""
                        userEmail = ""
                        userLocation = ""
                        screen = Screen.AuthLogin
                    },
                    themeMode = themeMode,
                    onThemeModeChange = { mode ->
                        themeMode = mode
                        SettingsStore.setThemeMode(mode)
                    }
                )


                Screen.OnboardingWelcome -> OnboardingWelcomeScreen {
                    screen = Screen.AuthSignupEmail
                }

                Screen.AuthSignupEmail -> AuthEmailScreen(
                    onContinue = { email ->

                        scope.launch {
                            val result = AuthRepository().sendOtp(email)

                            if (result.isSuccess) {
                                signupEmail = email
                                userEmail = email
                                screen = Screen.AuthOtp
                            } else {
                                println(
                                    "SEND OTP FAILED: " +
                                            result.exceptionOrNull()?.message
                                )
                            }
                        }
                    },

                    onBack = {
                        screen = Screen.OnboardingWelcome
                    },
                    onLogin = {
                        screen = Screen.AuthLogin
                    }
                )
                Screen.AuthOtp -> AuthOtpScreen(
                    email = signupEmail,

                    onVerify = { otp ->

                        scope.launch {

                            val result = AuthRepository().verifyOtp(
                                email = signupEmail,
                                otp = otp
                            )

                            if (result.isSuccess) {
                                signupOtp = otp
                                screen = Screen.AuthSignupDetails
                            } else {
                                println(
                                    "VERIFY OTP FAILED: " +
                                            result.exceptionOrNull()?.message
                                )
                            }
                        }
                    },

                    onResend = {

                        scope.launch {

                            val result =
                                AuthRepository().resendOtp(signupEmail)

                            if (!result.isSuccess) {
                                println(
                                    "RESEND OTP FAILED: " +
                                            result.exceptionOrNull()?.message
                                )
                            }
                        }
                    },

                    onLogin = {
                        screen = Screen.AuthLogin
                    },

                    onBack = {
                        screen = Screen.AuthSignupEmail
                    }
                )


                Screen.AuthLogin -> AuthLoginScreen(
                    onLogin = { email, password ->
                        scope.launch {
                            val result = AuthRepository().login(email, password)
                            if (result.isSuccess) {
                                userName = AuthStore.userName ?: userName
                                userEmail = AuthStore.userEmail ?: userEmail
                                screen = Screen.Feed
                            } else {
                                println("LOGIN FAILED: ${result.exceptionOrNull()?.message}")
                            }
                        }
                    },
                    onSignUp = {
                        screen = Screen.AuthSignupEmail
                    },

                    onBack = {
                        screen = Screen.AuthSignupEmail
                    }
                )

                Screen.AuthSignupDetails -> AuthSignupDetailsScreen(
                    email = signupEmail,

                    onContinue = { name, password ->

                        scope.launch {

                            val registerResult =
                                AuthRepository().register(
                                    name = name,
                                    email = signupEmail,
                                    password = password
                                )

                            if (registerResult.isSuccess) {

                                val loginResult =
                                    AuthRepository().login(
                                        email = signupEmail,
                                        password = password
                                    )

                                if (loginResult.isSuccess) {

                                    userName =
                                        AuthStore.userName ?: name

                                    userEmail =
                                        AuthStore.userEmail ?: signupEmail

                                    onboardingData =
                                        OnboardingData(name = name)

                                    screen =
                                        Screen.OnboardingBasicInfo

                                } else {

                                    println(
                                        "LOGIN AFTER SIGNUP FAILED: " +
                                                loginResult.exceptionOrNull()?.message
                                    )

                                    screen = Screen.AuthLogin
                                }

                            } else {

                                println(
                                    "REGISTER FAILED: " +
                                            registerResult.exceptionOrNull()?.message
                                )
                            }
                        }
                    },

                    onBack = {
                        screen = Screen.AuthOtp
                    }
                )
                Screen.OnboardingBasicInfo -> OnboardingBasicInfoScreen(
                    onNext = { description, role ->
                        selectedRole = role
                        onboardingData = onboardingData.copy(name = description, role = role)
                        screen = Screen.OnboardingLocation
                    },
                    onBack = { screen = Screen.OnboardingWelcome }
                )

                Screen.OnboardingLocation -> OnboardingLocationScreen(
                    onNext = { location ->
                        userLocation = location
                        onboardingData = onboardingData.copy(location = location)
                        screen = when (selectedRole) {
                            "Artist" -> Screen.ArtistExperience
                            "Event Organizer" -> Screen.OrganizerType
                            else -> Screen.AudienceInterests
                        }
                    },
                    onBack = { screen = Screen.OnboardingBasicInfo }
                )

                Screen.ArtistExperience -> ExperienceScreen(
                    onNext = { experience ->
                        onboardingData = onboardingData.copy(bio = experience)
                        screen = Screen.OnboardingDone
                    },
                    onBack = { screen = Screen.OnboardingLocation }
                )

                Screen.OrganizerType -> OrganizerTypeScreen(
                    onNext = { screen = Screen.OrganizerIntent },
                    onBack = { screen = Screen.OnboardingLocation }
                )

                Screen.OrganizerIntent -> OrganizerIntentScreen(
                    onNext = { screen = Screen.OnboardingDone },
                    onBack = { screen = Screen.OrganizerType }
                )

                Screen.AudienceInterests -> InterestsScreen(
                    onNext = { labels ->
                        onboardingData = onboardingData.copy(labels = labels)
                        screen = Screen.OnboardingDone
                    },
                    onBack = { screen = Screen.OnboardingLocation }
                )

                Screen.OnboardingDone -> OnboardingDoneScreen(
                    onFinish = {
                        scope.launch {
                            screen = when (selectedRole) {
                                "Artist" -> Screen.ArtistHome(userId = "123")
                                "Event Organizer" -> Screen.OrganizerHome(userId = "123")
                                else -> Screen.Marketplace
                            }
                            val response = onboardingRepository.onboardUser(
                                name = userName.ifBlank { onboardingData.name },
                                role = onboardingData.role,
                                location = onboardingData.location,
                                labels = onboardingData.labels,
                                bio = onboardingData.bio,
                                profilePicture = onboardingData.profilePicture
                            )
                            if (response.errors.isNullOrEmpty() && response.data?.onboardUser == true) {
                                // Success
                            } else {
                                println("ONBOARDING FAILED: ${response.errors}")
                            }
                        }
                    }
                )

                is Screen.Profile -> {
                    BackHandler { screen = Screen.Feed }
                    val presenter = remember(currentScreen.userId, currentProfile) {
                        ProfilePresenter(
                            repository = ProfileRepository(
                                initialProfile = currentProfile ?: Profile(
                                    id = currentScreen.userId,
                                    name = userName.ifBlank { AuthStore.userName.orEmpty() },
                                    email = userEmail.ifBlank { AuthStore.userEmail.orEmpty() },
                                    location = userLocation,
                                )
                            )
                        )
                    }
                    Scaffold(
                        bottomBar = {
                            KalaBottomNav(
                                selectedIndex = 3,
                                onStoreClick = { screen = Screen.Store },
                                onEventsClick = { screen = Screen.Events },
                                onHomeClick = { screen = Screen.Feed },
                                onProfileClick = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) }
                            )
                        },
                        floatingActionButton = {
                            FloatingActionButton(
                                onClick = { screen = Screen.UploadPost },
                                containerColor = BrandPurple,
                                contentColor = Color.White,
                                shape = CircleShape
                            ) {
                                Icon(Icons.Default.Add, contentDescription = "Add Post")
                            }
                        }
                    ) { innerPadding ->
                        Box(modifier = Modifier.padding(innerPadding)) {
                            ProfileScreen(
                                presenter = presenter,
                                userId = currentScreen.userId,
                                onEditProfile = { screen = Screen.EditProfile(currentScreen.userId) },
                                onShare = { /* Handle share */ },
                                posts = publishedPosts,
                                onBack = { screen = Screen.Feed }
                            )
                        }
                    }
                }

                Screen.UploadPost -> {
                    BackHandler { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) }
                    UploadPostScreen(
                        onDone = { desc, imgs ->
                            screen = Screen.PostPreview(
                                description = desc,
                                imageBytes = imgs,
                                userName = currentProfile?.name ?: userName,
                                userAvatarUrl = currentProfile?.avatarUrl,
                                userAvatarBytes = currentProfile?.avatarBytes
                            )
                        },
                        onBack = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) }
                    )
                }

                is Screen.PostPreview -> {
                    BackHandler { screen = Screen.UploadPost }
                    PostPreviewScreen(
                        description = currentScreen.description,
                        imageBytes = currentScreen.imageBytes,
                        userName = currentScreen.userName,
                        userAvatarUrl = currentScreen.userAvatarUrl,
                        userAvatarBytes = currentScreen.userAvatarBytes,
                        onPublish = {
                            scope.launch {
                                val published = PostRepository().createPost(currentScreen.description, currentScreen.imageBytes)
                                publishedPosts = publishedPosts + DraftPost(
                                    timeAgo = "Just now",
                                    content = currentScreen.description,
                                    likes = 0,
                                    comments = 0,
                                    hasImage = currentScreen.imageBytes.isNotEmpty(),
                                    imageBytes = currentScreen.imageBytes
                                )
                                screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString())
                            }
                        },
                        onBack = { screen = Screen.UploadPost }
                    )
                }

                is Screen.EditProfile -> {
                    BackHandler { screen = Screen.Profile(userId = currentScreen.userId) }
                    val profileToEdit = currentProfile ?: Profile(
                        id = currentScreen.userId,
                        name = userName.ifBlank { AuthStore.userName.orEmpty() },
                        location = userLocation,
                        username = "",
                        email = userEmail.ifBlank { AuthStore.userEmail.orEmpty() },
                    )
                    EditProfileScreen(
                        profile = profileToEdit,
                        onBack = { screen = Screen.Profile(userId = currentScreen.userId) },
                        onSave = { updated ->
                            currentProfile = updated
                            userName = updated.name
                            userEmail = updated.email
                            screen = Screen.Profile(userId = currentScreen.userId)
                        },
                    )
                }

                is Screen.ArtistHome -> {
                    BackHandler { screen = Screen.Feed }
                    ArtistHomeScreen(
                        viewModel = sharedEventListViewModel,
                        onEventClick = { eventId -> screen = Screen.EventDetails(eventId) },
                        onSwitchRole = { screen = Screen.OrganizerHome(userId = "123") },
                        onMenuClick = { scope.launch { drawerState.open() } },
                        onStoreClick = { screen = Screen.Store },
                        onMyEventsClick = { screen = Screen.OrganizerHome(userId = "123") },
                        onEventsClick = { screen = Screen.Events },
                        onHomeClick = { screen = Screen.Feed },
                        onProfileClick = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) }
                    )
                }

                Screen.Events -> {
                    BackHandler { screen = Screen.Feed }
                    ArtistHomeScreen(
                        viewModel = sharedEventListViewModel,
                        onEventClick = { eventId -> screen = Screen.EventDetails(eventId) },
                        onSwitchRole = { screen = Screen.OrganizerHome(userId = "123") },
                        onMenuClick = { scope.launch { drawerState.open() } },
                        onStoreClick = { screen = Screen.Store },
                        onMyEventsClick = { screen = Screen.OrganizerHome(userId = "123") },
                        onEventsClick = { screen = Screen.Events },
                        onHomeClick = { screen = Screen.Feed },
                        onProfileClick = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) }
                    )
                }

                is Screen.EventDetails -> {
                    BackHandler { screen = Screen.ArtistHome(userId = "123") }
                    EventDetailsScreen(
                        eventId = currentScreen.eventId,
                        viewModel = sharedEventListViewModel,
                        onBack = { screen = Screen.ArtistHome(userId = "123") },
                        onApplyOpportunity = { oppId, oppTitle ->
                            screen = Screen.ApplicationForm(currentScreen.eventId, oppId, oppTitle)
                        },
                    )
                }

                is Screen.ApplicationForm -> {
                    BackHandler { screen = Screen.EventDetails(currentScreen.eventId) }
                    val events by sharedEventListViewModel.events.collectAsState()
                    val event = events.firstOrNull { it.id == currentScreen.eventId }
                    ApplicationFormScreen(
                        eventId = currentScreen.eventId,
                        eventTitle = event?.title ?: "Event",
                        opportunityId = currentScreen.opportunityId,
                        opportunityTitle = currentScreen.opportunityTitle,
                        eventCoverBytes = event?.coverImageBytes,
                        applicantAvatarBytes = currentProfile?.avatarBytes,
                        onBack = { screen = Screen.EventDetails(currentScreen.eventId) },
                        onSubmit = { application: Application ->
                            ApplicationStore.addApplication(application)
                            screen = Screen.ApplicationStatus(application.id)
                        },
                    )
                }

                is Screen.ApplicationStatus -> {
                    BackHandler { screen = Screen.MyApplications(userId = "123") }
                    val app = ApplicationStore.applicationById(currentScreen.applicationId)
                    ApplicationStatusScreen(
                        eventTitle = app?.eventTitle ?: "Event",
                        onBack = { screen = Screen.MyApplications(userId = "123") },
                        onDone = { screen = Screen.MyApplications(userId = "123") },
                    )
                }

                is Screen.MyApplications -> {
                    BackHandler { screen = Screen.Feed }
                    MyApplicationsScreen(
                        onApplicationClick = { appId -> screen = Screen.ApplicationStatus(appId) },
                        onBack = { screen = Screen.Feed },
                        onSwitchRole = { screen = Screen.OrganizerHome(userId = "123") },
                        onMenuClick = { scope.launch { drawerState.open() } }
                    )
                }

                is Screen.OrganizerHome -> {
                    BackHandler { screen = Screen.Feed }

                    OrganizerHomeScreen(
                        userId = currentScreen.userId,
                        viewModel = sharedEventListViewModel,

                        onCreateEvent = {
                            draftEvent = EventDraft()
                            screen = Screen.CreateEvent
                        },

                        onMenuClick = {
                            scope.launch { drawerState.open() }
                        },

                        onEventClick = { eventId ->
                            screen = Screen.EventApplications(eventId)
                        },

                        onHomeClick = {
                            screen = Screen.Feed
                        },

                        onEventsClick = {
                            screen = Screen.Events
                        },

                        onStoreClick = {
                            screen = Screen.Store
                        },

                        onProfileClick = {
                            screen = Screen.Profile(
                                userId = (AuthStore.userId ?: 123).toString()
                            )
                        }
                    )
                }

                is Screen.OrganizerEventList -> {
                    BackHandler { screen = Screen.Feed }

                    OrganizerHomeScreen(
                        userId = currentScreen.userId,
                        viewModel = sharedEventListViewModel,

                        onCreateEvent = {
                            draftEvent = EventDraft()
                            screen = Screen.CreateEvent
                        },

                        onMenuClick = {
                            scope.launch { drawerState.open() }
                        },

                        onEventClick = { eventId ->
                            screen = Screen.EventApplications(eventId)
                        },

                        onHomeClick = {
                            screen = Screen.Feed
                        },

                        onEventsClick = {
                            screen = Screen.Events
                        },

                        onStoreClick = {
                            screen = Screen.Store
                        },

                        onProfileClick = {
                            screen = Screen.Profile(
                                userId = (AuthStore.userId ?: 123).toString()
                            )
                        }
                    )
                }

                Screen.CreateEvent -> {
                    BackHandler { screen = Screen.OrganizerHome(userId = "123") }
                    CreateEventScreen(
                        initialDraft = draftEvent,
                        onNext = { title, description, email, phone, coverBytes, galleryBytes ->
                            draftEvent = draftEvent.copy(title = title, description = description, email = email, phone = phone, coverImageBytes = coverBytes, galleryBytes = galleryBytes)
                            screen = Screen.SelectArtistCategories
                        },
                        onBack = { screen = Screen.OrganizerHome(userId = "123") },
                    )
                }

                Screen.SelectArtistCategories -> {
                    BackHandler { screen = Screen.CreateEvent }
                    SelectArtistCategoriesScreen(
                        onNext = { categories ->
                            draftEvent = draftEvent.copy(categories = categories)
                            screen = Screen.TimelineAndLocation
                        },
                        onBack = { screen = Screen.CreateEvent },
                    )
                }

                Screen.TimelineAndLocation -> {
                    BackHandler { screen = Screen.SelectArtistCategories }
                    TimelineAndLocationScreen(
                        onNext = { location, startDate, endDate ->
                            draftEvent = draftEvent.copy(location = location, startDate = startDate, endDate = endDate)
                            screen = Screen.ReviewEvent
                        },
                        onBack = { screen = Screen.SelectArtistCategories },
                    )
                }

                Screen.ReviewEvent -> {
                    BackHandler { screen = Screen.TimelineAndLocation }
                    ReviewEventScreen(
                        draft = draftEvent,
                        onBack = { screen = Screen.TimelineAndLocation },
                        onEdit = { screen = Screen.CreateEvent },
                        onPublish = {
                            val startDate = draftEvent.startDate
                            val duration = if (draftEvent.startDate != null && draftEvent.endDate != null) {
                                val days = draftEvent.endDate!!.toEpochDays() - draftEvent.startDate!!.toEpochDays() + 1
                                "$days days"
                            } else {
                                "1 day"
                            }
                            scope.launch {
                                try {
                                    val response = eventRepository.createEvent(draftEvent.title.trim().ifBlank { "Untitled Event" }, startDate?.toString() ?: "2025-10-24", duration)
                                    val createdEvent = response.data?.createEvent
                                    if (createdEvent != null && response.exception == null && response.errors.isNullOrEmpty()) {
                                        sharedEventListViewModel.addEvent(Event(id = createdEvent.id, title = createdEvent.name, description = draftEvent.description, location = draftEvent.location, startDate = draftEvent.startDate ?: LocalDate.parse("2025-10-24"), endDate = draftEvent.endDate, organizerName = "Organizer", email = draftEvent.email, phone = draftEvent.phone, coverImageBytes = draftEvent.coverImageBytes, galleryBytes = draftEvent.galleryBytes, categories = draftEvent.categories))
                                    } else {
                                        sharedEventListViewModel.addEvent(Event(id = (1000..9999).random().toString(), title = draftEvent.title.trim().ifBlank { "Untitled Event" }, description = draftEvent.description, location = draftEvent.location, startDate = draftEvent.startDate ?: LocalDate.parse("2025-10-24"), endDate = draftEvent.endDate, organizerName = "Organizer", email = draftEvent.email, phone = draftEvent.phone, coverImageBytes = draftEvent.coverImageBytes, galleryBytes = draftEvent.galleryBytes, categories = draftEvent.categories))
                                    }
                                } catch (e: Exception) {
                                    sharedEventListViewModel.addEvent(Event(id = (1000..9999).random().toString(), title = draftEvent.title.trim().ifBlank { "Untitled Event" }, description = draftEvent.description, location = draftEvent.location, startDate = draftEvent.startDate ?: LocalDate.parse("2025-10-24"), endDate = draftEvent.endDate, organizerName = "Organizer", email = draftEvent.email, phone = draftEvent.phone, coverImageBytes = draftEvent.coverImageBytes, galleryBytes = draftEvent.galleryBytes, categories = draftEvent.categories))
                                } finally {
                                    draftEvent = EventDraft()
                                    screen = Screen.OrganizerHome(userId = "123")
                                }
                            }
                        },
                    )
                }

                is Screen.EventApplications -> {
                    BackHandler { screen = Screen.OrganizerHome(userId = "123") }
                    val events by sharedEventListViewModel.events.collectAsState()
                    val event = events.firstOrNull { it.id == currentScreen.eventId }
                    if (event == null) {
                        LaunchedEffect(Unit) { screen = Screen.OrganizerHome(userId = "123") }
                    } else {
                        EventApplicationsScreen(
                            event = event,
                            onBack = { screen = Screen.OrganizerHome(userId = "123") },
                            onApplicationClick = { appId -> screen = Screen.ApplicationPreview(appId) },
                            onCreateOpportunity = { screen = Screen.CreateOpportunity(event.id) },
                            onOpportunityClick = { oppId -> screen = Screen.EditOpportunity(oppId, event.id) }
                        )
                    }
                }

                is Screen.CreateOpportunity -> {
                    BackHandler { screen = Screen.EventApplications(currentScreen.eventId) }
                    CreateOpportunityScreen(
                        eventId = currentScreen.eventId,
                        onBack = { screen = Screen.EventApplications(currentScreen.eventId) },
                        onFinish = { screen = Screen.EventApplications(currentScreen.eventId) }
                    )
                }

                is Screen.EditOpportunity -> {
                    BackHandler { screen = Screen.EventApplications(currentScreen.eventId) }
                    EditOpportunityScreen(
                        opportunityId = currentScreen.opportunityId,
                        onBack = { screen = Screen.EventApplications(currentScreen.eventId) },
                        onFinish = { screen = Screen.EventApplications(currentScreen.eventId) }
                    )
                }

                is Screen.ApplicationPreview -> {
                    val app = ApplicationStore.applicationById(currentScreen.applicationId)
                    if (app == null) {
                        LaunchedEffect(Unit) { screen = Screen.OrganizerHome(userId = "123") }
                    } else {
                        BackHandler { screen = Screen.EventApplications(app.eventId) }
                        ApplicationPreviewScreen(
                            application = app,
                            onBack = { screen = Screen.EventApplications(app.eventId) },
                            onAccept = { scope.launch { ApplicationRepository.updateStatus(app.id, ApplicationStatus.ACCEPTED) } },
                            onReject = { scope.launch { ApplicationRepository.updateStatus(app.id, ApplicationStatus.REJECTED) } },
                            onDone = { screen = Screen.EventApplications(app.eventId) },
                        )
                    }
                }
            }
        }
    }
}