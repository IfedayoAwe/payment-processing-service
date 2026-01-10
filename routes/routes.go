package routes

import (
	"net/http"

	"github.com/IfedayoAwe/payment-processing-service/config"
	"github.com/IfedayoAwe/payment-processing-service/handlers"
	"github.com/IfedayoAwe/payment-processing-service/middleware"
	echo "github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

func Register(e *echo.Echo, cfg *config.Config, handlers *handlers.Handlers) {
	// Health Check
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Ok")
	})

	// ReDoc UI - serves the OpenAPI spec via ReDoc
	e.GET("/docs", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
    <title>Payment Processing System API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <redoc spec-url='/docs/openapi.json'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`)
	})

	// OpenAPI JSON endpoint
	e.GET("/docs/openapi.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, getOpenAPISpec())
	})

	api := e.Group("/api")
	api.Use(emw.Logger(), emw.Recover())

	// Protected Routes - require X-User-ID header
	api.Use(middleware.UserIDMiddleware())

	// Register payment routes
	RegisterPaymentRoutes(api, handlers)
}

func RegisterPaymentRoutes(api *echo.Group, handlers *handlers.Handlers) {
	// Payment routes will be added here
	// Examples:
	// api.POST("/payments/internal", handlers.Payment.CreateInternalPayment)
	// api.POST("/payments/external", handlers.Payment.CreateExternalPayment)
	// api.GET("/payments/:id", handlers.Payment.GetPayment)
	// api.GET("/payments", handlers.Payment.ListPayments)
}

func getOpenAPISpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Payment Processing System API",
			"version":     "1.0.0",
			"description": "Multi-currency payment processing system API. Handles payments between internal users and to external accounts in USD, EUR, and GBP.",
		},
		"servers": []map[string]interface{}{
			{
				"url":         "http://localhost:8080",
				"description": "Local development server",
			},
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check",
					"description": "Check if the service is running",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
							"content": map[string]interface{}{
								"text/plain": map[string]interface{}{
									"schema": map[string]interface{}{
										"type":    "string",
										"example": "Ok",
									},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"X-User-ID": map[string]interface{}{
					"type":        "apiKey",
					"in":          "header",
					"name":        "X-User-ID",
					"description": "User ID for authentication (e.g., user_1)",
				},
			},
		},
	}
}
