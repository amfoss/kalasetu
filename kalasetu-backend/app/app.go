package app

import (
	"database/sql"
	"kalasetu/config"
	"kalasetu/graph"
	"kalasetu/handlers"
	"kalasetu/middlewares"
	"kalasetu/migrations"
	"kalasetu/payments"
	"kalasetu/repos"
	"kalasetu/routes"
	"kalasetu/services"
	"log"
	"os"

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
	// Payments is not consumed until checkout lands; it is held here so the
	// marketplace services can be wired with it.
	Payments payments.PaymentProvider
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

	app := New(db, payments.NewUnconfigured())
	app.Port = port
	return app
}

// New builds the App from an already-connected (and migrated) database and a
// PaymentProvider. It touches neither the environment nor the network, so
// tests can construct it in-process and drive app.Router directly.
func New(db *sql.DB, paymentProvider payments.PaymentProvider) *App {
	r := gin.Default()

	userRepo := repos.NewUserRepository(db)
	refreshTokenRepo := repos.NewRefreshTokenRepository(db)
	authService := services.NewAuthService(userRepo, refreshTokenRepo)
	authHandler := handlers.NewAuthHandler(authService)

	userService := services.NewUserService(userRepo)

	eventRepo := repos.NewEventRepository(db)
	eventService := services.NewEventService(eventRepo)

	applicationRepo := repos.NewApplicationRepository(db)
	applicationService := services.NewApplicationService(applicationRepo)

	listingRepo := repos.NewListingRepository(db)
	listingService := services.NewListingService(listingRepo)

	cartRepo := repos.NewCartRepository(db)
	cartService := services.NewCartService(cartRepo, listingRepo)

	apiV1 := r.Group("/api/v1")
	routes.RegisterAuthRoutes(apiV1, authHandler)

	resolver := graph.NewResolver(eventService, userService, applicationService, listingService, cartService)
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

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return srv
}
