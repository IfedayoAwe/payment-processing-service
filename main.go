package main

import (
	"log"

	"github.com/IfedayoAwe/payment-processing-service/config"
	"github.com/IfedayoAwe/payment-processing-service/db"
	"github.com/IfedayoAwe/payment-processing-service/handlers"
	"github.com/IfedayoAwe/payment-processing-service/routes"
	service "github.com/IfedayoAwe/payment-processing-service/services"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/go-playground/validator/v10"

	echo "github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func run() error {
	// Load config
	cfg := config.Load()

	// Initialize DB and queries
	queries, dbConn := db.InitDBWithDeps(cfg, db.DefaultDependencies)

	// Initialize Redis (cache)
	cache := setupRedis(&cfg)
	defer func() {
		if err := cache.Close(); err != nil {
			log.Printf("Error closing cache: %v", err)
		}
	}()

	// Init services and handlers
	services := service.NewServices(dbConn, queries, &cfg, cache)
	newHandlers := handlers.NewHandlers(services)

	// Init Echo
	e := echo.New()

	// Initialize validator with translations
	validate := utils.InitValidator()
	e.Validator = &CustomValidator{validator: validate}

	// Custom HTTP error handler to handle validation errors
	e.HTTPErrorHandler = utils.HTTPErrorHandler

	// Register routes
	routes.Register(e, &cfg, newHandlers)

	// Determine port
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Payment Processing Service running on port %s", port)

	return e.Start(":" + port)
}

func setupRedis(cfg *config.Config) utils.Cache {
	log.Println("Initializing Redis cache...", cfg.RedisURL)
	redis, err := utils.NewRedis(cfg)
	if err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}

	return redis
}
