package app

import (
	"database/sql"
	"kalasetu/config"
	"kalasetu/graph"
	"kalasetu/handlers"
	"kalasetu/middlewares"
	"kalasetu/migrations"
	"kalasetu/payments"
	"kalasetu/payments/razorpay"
	"kalasetu/repos"
	"kalasetu/routes"
	"kalasetu/services"
	"kalasetu/storage"
	"log"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

type App struct {
	Router *gin.Engine
	Srv    *handler.Server
	Port   string
	// Payments is the provider checkout charges through.
	Payments payments.PaymentProvider
}

// GatewayOptions configures the two-phase Checkout Session flow: Gateway is
// the payments.Gateway implementation to wire in, KeyID is its public key
// (handed to Buyers' browsers as-is), and ReservationWindow is how long
// Stock stays reserved by an open Checkout Session.
type GatewayOptions struct {
	Gateway           payments.Gateway
	KeyID             string
	ReservationWindow time.Duration
}

const defaultPort = "8080"

// NewApp is the production wiring: it reads configuration from the environment,
// connects to and migrates the database, and builds the App around it.
func NewApp() *App {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or failed to load. Falling back to system environment variables.")
	}

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

	razorpayCfg := config.LoadRazorpayConfig()
	var gateway payments.Gateway
	if razorpayCfg.IsConfigured() {
		gateway = razorpay.NewConnector(razorpayCfg, nil)
		log.Println("Razorpay gateway configured")
	} else {
		gateway = payments.NewUnconfiguredGateway()
		log.Println("Note: Razorpay (RAZORPAY_KEY_ID) is not configured. Checkout sessions cannot be created.")
	}

	app := New(db, payments.NewUnconfigured(), GatewayOptions{
		Gateway:           gateway,
		KeyID:             razorpayCfg.KeyID,
		ReservationWindow: razorpayCfg.ReservationWindow,
	})
	app.Port = port
	return app
}

// New builds the App from an already-connected (and migrated) database, a
// PaymentProvider and GatewayOptions. It touches neither the environment nor
// the network, so tests can construct it in-process and drive app.Router
// directly.
func New(db *sql.DB, paymentProvider payments.PaymentProvider, gatewayOpts GatewayOptions) *App {
	r := gin.Default()

	userRepo := repos.NewUserRepository(db)
	refreshTokenRepo := repos.NewRefreshTokenRepository(db)
	authService := services.NewAuthService(userRepo, refreshTokenRepo)
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

	listingRepo := repos.NewListingRepository(db)
	listingService := services.NewListingService(listingRepo)

	cartRepo := repos.NewCartRepository(db)
	cartService := services.NewCartService(cartRepo, listingRepo)

	orderRepo := repos.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, gatewayOpts.Gateway)

	checkoutSessionRepo := repos.NewCheckoutSessionRepository(db)
	checkoutSessionService := services.NewCheckoutSessionService(checkoutSessionRepo, gatewayOpts.Gateway, gatewayOpts.KeyID, gatewayOpts.ReservationWindow)

	apiV1 := r.Group("/api/v1")
	routes.RegisterAuthRoutes(apiV1, authHandler)

	resolver := graph.NewResolver(eventService, applicationService, opportunityService, postService, commentService, likeService, userService, profileService, listingService, cartService, orderService, checkoutSessionService)
	srv := gqlSetup(resolver)

	r.POST("/api/v1/graphql", middlewares.OptionalJWT(), func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL", "/api/v1/graphql").ServeHTTP(c.Writer, c.Request)
	})

	return &App{Router: r, Srv: srv, Port: defaultPort, Payments: paymentProvider}
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
