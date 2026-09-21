package com.example.kalasetu

import androidx.compose.foundation.layout.*
import androidx.compose.runtime.*
import androidx.compose.material3.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.ui.unit.dp
import com.example.kalasetu.features.auth.*
import com.example.kalasetu.features.feed.FeedScreen
import com.example.kalasetu.features.feed.KalaTopBar
import com.example.kalasetu.features.feed.KalaBottomNav
import com.example.kalasetu.features.feed.SidebarContent
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.kalasetu.features.application.*
import com.example.kalasetu.features.event.*
import com.example.kalasetu.features.opportunity.CreateOpportunityScreen
import com.example.kalasetu.features.opportunity.EditOpportunityScreen
import com.example.kalasetu.features.onboarding.*
import com.example.kalasetu.features.profile.*
import com.example.kalasetu.navigation.BackHandler
import com.example.kalasetu.navigation.Screen
import com.example.kalasetu.theme.KalasetuTheme
import kotlinx.datetime.LocalDate
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.kalasetu.features.feed.FeedViewModel
import com.example.kalasetu.repository.EventRepository
import kotlinx.coroutines.launch
import com.example.kalasetu.repository.AuthRepository
import com.example.kalasetu.repository.PostRepository
import com.example.kalasetu.repository.OnboardingRepository
import com.example.kalasetu.features.onboarding.OnboardingData
import kotlinx.coroutines.launch
import kotlin.random.Random

@Composable
fun App() {
    var screen by remember { mutableStateOf<Screen>(Screen.OnboardingWelcome) }
    var selectedRole by remember { mutableStateOf("") }
    var userName by remember { mutableStateOf("") }
    var userEmail by remember { mutableStateOf("") }
    var userLocation by remember { mutableStateOf("") }
    var currentProfile by remember { mutableStateOf<Profile?>(null) }
    var draftEvent by remember { mutableStateOf(EventDraft()) }
    var publishedPosts by remember { mutableStateOf<List<DraftPost>>(emptyList()) }
    val sharedEventListViewModel: EventListViewModel = viewModel()
    val eventRepository = remember { EventRepository() }
    var onboardingData by remember { mutableStateOf(OnboardingData()) }
    val onboardingRepository = remember { OnboardingRepository() }
    val drawerState = rememberDrawerState(initialValue = DrawerValue.Closed)
    val scope = rememberCoroutineScope()
    val feedViewModel: FeedViewModel = viewModel()
    val isAuthScreen = screen is Screen.OnboardingWelcome ||
                      screen is Screen.AuthSignup ||
                      screen is Screen.AuthOtp ||
                      screen is Screen.AuthLogin

    KalasetuTheme {
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
                        screen = Screen.Profile(
                            userId = AuthStore.userId?.toString() ?: ""
                        )
                    },
                    onNavigateToStore = { screen = Screen.Store },
                    onNavigateToHome = { screen = Screen.Feed },
                    onMenuClick = {
                        scope.launch {
                            drawerState.open()
                        }
                    }
                )
                is Screen.OrganizerEventList -> OrganizerHomeScreen(
                    userId = currentScreen.userId,
                    viewModel = sharedEventListViewModel,
                    onCreateEvent = {
                        draftEvent = EventDraft()
                        screen = Screen.CreateEvent
                    },
                    onMenuClick = {
                        scope.launch {
                            drawerState.open()
                        }
                    },
                    onEventClick = { eventId ->
                        screen = Screen.EventApplications(eventId)
                    },
                )
                Screen.Store -> {
                    BackHandler { screen = Screen.Feed }
                    Scaffold(
                        topBar = {
                            KalaTopBar(
                                avatarUrl = currentProfile?.avatarUrl,
                                avatarBytes = currentProfile?.avatarBytes,
                                userName = currentProfile?.name ?: userName,
                                onProfileClick = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) },
                                onMenuClick = { scope.launch { drawerState.open() } }
                            )
                        },
                        bottomBar = {
                            KalaBottomNav(
                                selectedIndex = 0,
                                onStoreClick = { screen = Screen.Store },
                                onHomeClick = { screen = Screen.Feed },
                                onProfileClick = { screen = Screen.Profile(userId = (AuthStore.userId ?: 123).toString()) }
                            )
                        }
                    ) { innerPadding ->
                        Box(Modifier.fillMaxSize().padding(innerPadding), contentAlignment = Alignment.Center) {
                            Text("Welcome to MarketPlace. The features will soon be added.!")

                        }
                    }
                }


            Screen.OnboardingWelcome -> OnboardingWelcomeScreen {
                screen = Screen.AuthSignup
            }
            Screen.AuthSignup -> AuthSignupScreen(
                onSignUp = { name, email, password ->
                    scope.launch {

                        val registerResult = AuthRepository().register(
                            name = name,
                            email = email,
                            password = password
                        )

                        if (registerResult.isSuccess) {

                            val loginResult = AuthRepository().login(
                                email = email,
                                password = password
                            )

                            if (loginResult.isSuccess) {

                                userName = AuthStore.userName ?: name
                                userEmail = AuthStore.userEmail ?: email

                                onboardingData = OnboardingData(
                                    name = name
                                )

                                screen = Screen.OnboardingBasicInfo

                            } else {
                                screen = Screen.AuthLogin
                            }
                        }
                    }
                },

                onLogin = {
                    screen = Screen.AuthLogin
                },

                onBack = {
                    screen = Screen.AuthLogin
                },
            )
            Screen.AuthOtp -> AuthOtpScreen(
                onVerify = { screen = Screen.OnboardingBasicInfo },
                onLogin = { screen = Screen.AuthLogin },
                onBack = { screen = Screen.AuthSignup },
            )
            Screen.AuthLogin -> AuthLoginScreen(
                onLogin = { email, password ->
                    scope.launch {

                        val result = AuthRepository().login(
                            email = email,
                            password = password
                        )

                        if (result.isSuccess) {

                            println("========== APP LOGIN SUCCESS ==========")

                            userName = AuthStore.userName ?: userName
                            userEmail = AuthStore.userEmail ?: userEmail

                            screen = Screen.Feed

                        } else {

                            println(
                                "LOGIN FAILED: ${
                                    result.exceptionOrNull()?.message
                                }"
                            )
                        }
                    }
                },

                onSignUp = { screen = Screen.AuthSignup },
                onBack = { screen = Screen.AuthSignup },
            )
            Screen.OnboardingBasicInfo -> OnboardingBasicInfoScreen(
                onNext = { description, role ->
                    selectedRole = role
                    onboardingData = onboardingData.copy(
                        name = description,
                        role = role
                    )
                    screen = Screen.OnboardingLocation
                },
                onBack = { screen = Screen.OnboardingWelcome }
            )
            Screen.OnboardingLocation -> OnboardingLocationScreen(
                onNext = { location ->
                    userLocation = location
                    onboardingData = onboardingData.copy(
                        location = location
                    )
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
            Screen.OrganizerType -> OrganizerTypeScreen( onNext = {screen = Screen.OrganizerIntent}, onBack = { screen = Screen.OnboardingLocation })
            Screen.OrganizerIntent -> OrganizerIntentScreen(onNext = { screen = Screen.OnboardingDone }) { screen = Screen.OrganizerType }
            Screen.AudienceInterests -> InterestsScreen(onNext = { screen = Screen.OnboardingDone }) { screen = Screen.OnboardingLocation }

            Screen.OnboardingDone -> {
                OnboardingDoneScreen(
                    onFinish = {
                        scope.launch {
                            screen = Screen.Feed
                            val response = onboardingRepository.onboardUser(
                                name = userName.ifBlank { onboardingData.name },
                                role = onboardingData.role,
                                location = onboardingData.location,
                                labels = onboardingData.labels,
                                bio = onboardingData.bio,
                                profilePicture = onboardingData.profilePicture
                            )

                            if (
                                response.errors.isNullOrEmpty() &&
                                response.data?.onboardUser == true
                            ) {

                            } else {
                                println("ONBOARDING FAILED: ${response.errors}")
                            }
                        }
                    }
                )
            }

                // ─── Profile ───
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
                                selectedIndex = 2,
                                onStoreClick = { screen = Screen.Store },
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
                    PostPreviewScreen(
                        description = currentScreen.description,
                        imageBytes = currentScreen.imageBytes,
                        userName = currentScreen.userName,
                        userAvatarUrl = currentScreen.userAvatarUrl,
                        userAvatarBytes = currentScreen.userAvatarBytes,
onPublish = {
                            scope.launch {
                                val published = PostRepository().createPost(
                                    content = currentScreen.description,
                                    images = currentScreen.imageBytes
                                )

                                println("========== DISPATCH PUBLISH ==========")
                                println("published = $published")

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

                // Artist Flow
                is Screen.ArtistHome -> ArtistHomeScreen(
                    viewModel = sharedEventListViewModel,
                    onEventClick = { eventId ->
                        screen = Screen.EventDetails(eventId)
                    },
                    onSwitchRole = {
                        screen = Screen.OrganizerHome(userId = "123")
                    },
                    onMenuClick = {
                        scope.launch {
                            drawerState.open()
                        }
                    }
                )

                is Screen.EventDetails -> EventDetailsScreen(
                    eventId = currentScreen.eventId,
                    viewModel = sharedEventListViewModel,
                    onBack = { screen = Screen.ArtistHome(userId = "123") },
                    onApplyOpportunity = { oppId, oppTitle ->
                        screen = Screen.ApplicationForm(
                            eventId = currentScreen.eventId,
                            opportunityId = oppId,
                            opportunityTitle = oppTitle
                        )
                    },
                )

                is Screen.ApplicationForm -> {
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
                        onApplicationClick = { appId ->
                            screen = Screen.ApplicationStatus(appId)
                        },
                        onBack = {
                            screen = Screen.Feed
                        },
                        onSwitchRole = {
                            screen = Screen.OrganizerHome(userId = "123")
                        },
                        onMenuClick = {
                            scope.launch {
                                drawerState.open()
                            }
                        }
                    )
                }

                // ─── Organizer Flow ───
                is Screen.OrganizerHome -> OrganizerHomeScreen(
                    userId = currentScreen.userId,
                    viewModel = sharedEventListViewModel,

                    onCreateEvent = {
                        draftEvent = EventDraft()
                        screen = Screen.CreateEvent
                    },

                    onMenuClick = {
                        scope.launch {
                            drawerState.open()
                        }
                    },

                    onEventClick = { eventId ->
                        screen = Screen.EventApplications(eventId)
                    },

                )

                Screen.CreateEvent -> CreateEventScreen(
                    initialDraft = draftEvent,
                    onNext = { title, description, email, phone, coverBytes, galleryBytes ->
                        draftEvent = draftEvent.copy(
                            title = title,
                            description = description,
                            email = email,
                            phone = phone,
                            coverImageBytes = coverBytes,
                            galleryBytes = galleryBytes,
                        )
                        screen = Screen.SelectArtistCategories
                    },
                    onBack = { screen = Screen.OrganizerHome(userId = "123") },
                )

                Screen.SelectArtistCategories -> SelectArtistCategoriesScreen(
                    onNext = { categories ->
                        draftEvent = draftEvent.copy(categories = categories)
                        screen = Screen.TimelineAndLocation
                    },
                    onBack = { screen = Screen.CreateEvent },
                )

                Screen.TimelineAndLocation -> TimelineAndLocationScreen(
                    onNext = { location, startDate, endDate ->
                        draftEvent = draftEvent.copy(
                            location = location,
                            startDate = startDate,
                            endDate = endDate,
                        )
                        screen = Screen.ReviewEvent
                    },
                    onBack = { screen = Screen.SelectArtistCategories },
                )

            Screen.ReviewEvent -> ReviewEventScreen(
                draft = draftEvent,
                onBack = { screen = Screen.TimelineAndLocation },
                onEdit = { screen = Screen.CreateEvent },
                onPublish = {
                    println("========== PUBLISH CLICKED ==========")
                    val startDate = draftEvent.startDate

                    val duration = if (draftEvent.startDate != null && draftEvent.endDate != null) {
                        val startDate = draftEvent.startDate!!
                        val endDate = draftEvent.endDate!!

                        val days = endDate.toEpochDays() - startDate.toEpochDays() + 1
                        "$days days"
                    } else {
                        "1 day"
                    }

                    scope.launch {
                        try {
                            println("========== CALLING GRAPHQL ==========")
                            val response = eventRepository.createEvent(
                                name = draftEvent.title.trim().ifBlank { "Untitled Event" },
                                startDate = startDate?.toString() ?: "2025-10-24",
                                duration = duration
                            )

                            val createdEvent = response.data?.createEvent
                            if (createdEvent != null && response.exception == null && response.errors.isNullOrEmpty()) {
                                println("========== EVENT CREATED SUCCESSFULLY ==========")
                                sharedEventListViewModel.addEvent(
                                    Event(
                                        id = createdEvent.id,
                                        title = createdEvent.name,
                                        description = draftEvent.description,
                                        location = draftEvent.location,
                                        startDate = draftEvent.startDate ?: kotlinx.datetime.LocalDate.parse("2025-10-24"),
                                        endDate = draftEvent.endDate,
                                        organizerName = "Organizer",
                                        email = draftEvent.email,
                                        phone = draftEvent.phone,
                                        coverImageBytes = draftEvent.coverImageBytes,
                                        galleryBytes = draftEvent.galleryBytes,
                                        categories = draftEvent.categories
                                    )
                                )
                            } else {
                                println("========== EVENT CREATION FALLBACK ==========")
                                println("data = ${response.data}, errors = ${response.errors}")
                                val fallbackId = (1000..9999).random().toString()
                                sharedEventListViewModel.addEvent(
                                    Event(
                                        id = fallbackId,
                                        title = draftEvent.title.trim().ifBlank { "Untitled Event" },
                                        description = draftEvent.description,
                                        location = draftEvent.location,
                                        startDate = draftEvent.startDate ?: kotlinx.datetime.LocalDate.parse("2025-10-24"),
                                        endDate = draftEvent.endDate,
                                        organizerName = "Organizer",
                                        email = draftEvent.email,
                                        phone = draftEvent.phone,
                                        coverImageBytes = draftEvent.coverImageBytes,
                                        galleryBytes = draftEvent.galleryBytes,
                                        categories = draftEvent.categories
                                    )
                                )
                            }
                        } catch (e: Exception) {
                            println("NETWORK ERROR: ${e.message}")
                            e.printStackTrace()
                            val fallbackId = (1000..9999).random().toString()
                            sharedEventListViewModel.addEvent(
                                Event(
                                    id = fallbackId,
                                    title = draftEvent.title.trim().ifBlank { "Untitled Event" },
                                    description = draftEvent.description,
                                    location = draftEvent.location,
                                    startDate = draftEvent.startDate ?: kotlinx.datetime.LocalDate.parse("2025-10-24"),
                                    endDate = draftEvent.endDate,
                                    organizerName = "Organizer",
                                    email = draftEvent.email,
                                    phone = draftEvent.phone,
                                    coverImageBytes = draftEvent.coverImageBytes,
                                    galleryBytes = draftEvent.galleryBytes,
                                    categories = draftEvent.categories
                                )
                            )
                        } finally {
                            draftEvent = EventDraft()
                            screen = Screen.OrganizerHome(userId = "123")
                        }
                    }
                },
            )

                // ─── Event Applications (Organizer) ───
                is Screen.EventApplications -> {
                    val events by sharedEventListViewModel.events.collectAsState()
                    val event = events.firstOrNull { it.id == currentScreen.eventId }
                    if (event == null) {
                        LaunchedEffect(Unit) {
                            screen = Screen.OrganizerHome(userId = "123")
                        }
                    } else {
                        EventApplicationsScreen(
                            event = event,
                            onBack = { screen = Screen.OrganizerHome(userId = "123") },
                            onApplicationClick = { appId ->
                                screen = Screen.ApplicationPreview(appId)
                            },
                            onCreateOpportunity = {
                                screen = Screen.CreateOpportunity(event.id)
                            },
                            onOpportunityClick = { oppId ->
                                screen = Screen.EditOpportunity(oppId, event.id)
                            }
                        )
                    }
                }

                // ─── Create Opportunity (Organizer) ───
                is Screen.CreateOpportunity -> {
                    CreateOpportunityScreen(
                        eventId = currentScreen.eventId,
                        onBack = { screen = Screen.EventApplications(currentScreen.eventId) },
                        onFinish = { screen = Screen.EventApplications(currentScreen.eventId) }
                    )
                }

                // ─── Edit Opportunity (Organizer) ───
                is Screen.EditOpportunity -> {
                    EditOpportunityScreen(
                        opportunityId = currentScreen.opportunityId,
                        onBack = { screen = Screen.EventApplications(currentScreen.eventId) },
                        onFinish = { screen = Screen.EventApplications(currentScreen.eventId) }
                    )
                }

                // ─── Application Preview (Organizer) ───
                is Screen.ApplicationPreview -> {
                    val app = ApplicationStore.applicationById(currentScreen.applicationId)
                    if (app == null) {
                        LaunchedEffect(Unit) {
                            screen = Screen.OrganizerHome(userId = "123")
                        }
                    } else {
                        ApplicationPreviewScreen(
                            application = app,
                            onBack = { screen = Screen.EventApplications(app.eventId) },
                            onAccept = {
                                scope.launch {
                                    ApplicationRepository.updateStatus(app.id, ApplicationStatus.ACCEPTED)
                                }
                            },
                            onReject = {
                                scope.launch {
                                    ApplicationRepository.updateStatus(app.id, ApplicationStatus.REJECTED)
                                }
                            },
                            onDone = { screen = Screen.EventApplications(app.eventId) },
                        )
                    }
                }
            }
        }
    }
}