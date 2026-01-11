package main

import (
	"context"

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

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func main() {
	utils.InitLogger()
	logger := utils.Logger

	if err := run(); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}
}

func run() error {
	logger := utils.Logger
	cfg := config.Load()

	queries, dbConn := db.InitDBWithDeps(cfg, db.DefaultDependencies)

	redisClient, err := setupRedis(&cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize redis")
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error().Err(err).Msg("error closing redis")
		}
	}()

	services := service.NewServices(dbConn, queries, &cfg, redisClient.GetClient())
	newHandlers := handlers.NewHandlers(services)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	services.StartWorkers(ctx)

	e := echo.New()

	validate := utils.InitValidator()
	e.Validator = &CustomValidator{validator: validate}
	e.HTTPErrorHandler = utils.HTTPErrorHandler

	routes.Register(e, &cfg, newHandlers)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	logger.Info().Str("port", port).Msg("Payment Processing Service running")

	return e.Start(":" + port)
}

func setupRedis(cfg *config.Config) (*utils.Redis, error) {
	return utils.NewRedis(cfg)
}
