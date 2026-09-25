package app

import (
	"kalasetu/config"
	"kalasetu/graph"
	"kalasetu/handlers"
	"kalasetu/migrations"
	"kalasetu/repos"
	"kalasetu/routes"
	"kalasetu/services"
	"kalasetu/storage"
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

type App struct {
	Router *gin.Engine
	Srv    *handler.Server
	Port   string
}

const defaultPort = "8080"

func NewApp() *App {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or failed to load. Falling back to system environment variables.")
	}

	r := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db, err := config.InitDB()
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v. Database operations will fail at runtime.", err)
	} else {
		// Run database migrations
		if err := migrations.RunMigrations(db); err != nil {
			log.Printf("Warning: Failed to run database migrations: %v", err)
		}
		log.Printf("Migrations done")
	}

	emailCfg := config.LoadEmailConfig()
	emailService := services.NewEmailService(emailCfg)

	otpRepo := repos.NewOTPRepository(db)
	userRepo := repos.NewUserRepository(db)
	refreshTokenRepo := repos.NewRefreshTokenRepository(db)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, otpRepo, emailService)
	authHandler := handlers.NewAuthHandler(authService)

	userService := services.NewUserService(userRepo)

	var objectStorage storage.ObjectStorage
	storageCfg := config.LoadStorageConfig()
	if storageCfg.IsConfigured() {
		s3Storage, err := storage.NewS3(storageCfg)
		if err != nil {
			log.Printf("Warning: failed to initialise object storage: %v. Post media and event banner uploads will fail at runtime.", err)
		} else {
			objectStorage = s3Storage
			log.Printf("Object storage configured for bucket %q in region %q", storageCfg.Bucket, storageCfg.Region)
		}
	} else {
		log.Println("Note: object storage (AWS_BUCKET) is not configured. Posts and events can be created without media.")
	}

	eventRepo := repos.NewEventRepository(db)
	eventService := services.NewEventService(eventRepo, objectStorage)

	applicationRepo := repos.NewApplicationRepository(db)
	applicationService := services.NewApplicationService(applicationRepo)

	opportunityRepo := repos.NewOpportunityRepository(db)
	opportunityService := services.NewOpportunityService(opportunityRepo)

	postRepo := repos.NewPostRepository(db)
	postMediaRepo := repos.NewPostMediaRepository(db)

	postService := services.NewPostService(postRepo, postMediaRepo, objectStorage)

	commentRepo := repos.NewCommentRepository(db)
	commentService := services.NewCommentService(commentRepo)

	likeRepo := repos.NewLikeRepository(db)
	likeService := services.NewLikeService(likeRepo)

	profileRepo := repos.NewProfileRepository(db)
	profileService := services.NewProfileService(profileRepo, objectStorage)

	apiV1 := r.Group("/api/v1")
	routes.RegisterAuthRoutes(apiV1, authHandler)

	resolver := graph.NewResolver(eventService, applicationService, opportunityService, postService, commentService, likeService, userService, profileService)
	srv := gqlSetup(resolver)

	return &App{Router: r, Srv: srv, Port: port}
}

func gqlSetup(resolver *graph.Resolver) *handler.Server {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return srv
}
