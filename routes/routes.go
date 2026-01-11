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
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Ok")
	})

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

	e.GET("/docs/openapi.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, getOpenAPISpec())
	})

	api := e.Group("/api")
	api.Use(emw.Logger(), emw.Recover())
	api.Use(middleware.UserIDMiddleware())

	RegisterPaymentRoutes(api, handlers)

	// Test endpoint (no authentication required)
	testApi := e.Group("/api/test")
	testApi.Use(emw.Logger(), emw.Recover())
	RegisterTestRoutes(testApi, handlers)
}

func RegisterPaymentRoutes(api *echo.Group, handlers *handlers.Handlers) {
	paymentHandler := handlers.Payment()
	webhookHandler := handlers.Webhook()
	nameEnquiryHandler := handlers.NameEnquiry()

	api.GET("/exchange-rate", paymentHandler.GetExchangeRate)
	api.GET("/wallets", paymentHandler.GetUserWallets)
	api.GET("/transactions", paymentHandler.GetTransactionHistory)
	api.POST("/payments/internal", paymentHandler.CreateInternalTransfer)
	api.POST("/payments/external", paymentHandler.CreateExternalTransfer)
	api.POST("/payments/:id/confirm", paymentHandler.ConfirmTransaction)
	api.GET("/payments/:id", paymentHandler.GetTransaction)

	api.POST("/webhooks/:provider", webhookHandler.ReceiveWebhook)

	api.POST("/name-enquiry", nameEnquiryHandler.EnquireAccountName)
}

func RegisterTestRoutes(testApi *echo.Group, handlers *handlers.Handlers) {
	paymentHandler := handlers.Payment()
	testApi.GET("/users", paymentHandler.GetTestUsers)
}

func getOpenAPISpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Payment Processing System API",
			"version":     "1.0.0",
			"description": "Multi-currency payment processing system API. Handles payments between internal users and to external accounts in USD, EUR, and GBP. Supports exchange rate locking, double-entry ledger accounting, and asynchronous processing for external transfers.",
		},
		"servers": []map[string]interface{}{
			{
				"url":         "http://localhost:8080",
				"description": "Local development server",
			},
		},
		"paths": map[string]interface{}{
			"/health":                    getHealthEndpoint(),
			"/api/test/users":            getTestUsersEndpoint(),
			"/api/name-enquiry":          getNameEnquiryEndpoint(),
			"/api/exchange-rate":         getExchangeRateEndpoint(),
			"/api/wallets":               getWalletsEndpoint(),
			"/api/payments/internal":     getCreateInternalTransferEndpoint(),
			"/api/payments/external":     getCreateExternalTransferEndpoint(),
			"/api/payments/{id}/confirm": getConfirmTransactionEndpoint(),
			"/api/payments/{id}":         getGetTransactionEndpoint(),
			"/api/transactions":          getTransactionHistoryEndpoint(),
			"/api/webhooks/{provider}":   getWebhookEndpoint(),
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
			"schemas": getSchemas(),
		},
	}
}

func getHealthEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Health check",
			"description": "Check if the service is running",
			"operationId": "healthCheck",
			"tags":        []string{"Health"},
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
	}
}

func getTestUsersEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get test users",
			"description": "Returns test user data including user IDs, PINs, wallets, bank accounts, and balances. No authentication required. Useful for testing and evaluation.",
			"operationId": "getTestUsers",
			"tags":        []string{"Test"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Test users retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/TestUsersResponse",
							},
							"example": map[string]interface{}{
								"data": map[string]interface{}{
									"users": []map[string]interface{}{
										{
											"user_id": "user_1",
											"name":    "John Doe",
											"pin":     "12345",
											"wallets": []map[string]interface{}{
												{
													"id":             "wallet_user1_usd",
													"currency":       "USD",
													"balance":        100.50,
													"account_number": "1000000001",
													"bank_name":      "Test Bank",
													"bank_code":      "044",
													"account_name":   "John Doe",
													"provider":       "currencycloud",
												},
											},
										},
									},
								},
								"message": "test users retrieved successfully",
							},
						},
					},
				},
			},
		},
	}
}

func getNameEnquiryEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Name enquiry",
			"description": "Check if an account number and bank code belong to an internal user or external account. Returns account name, whether it's internal, and currency.",
			"operationId": "nameEnquiry",
			"tags":        []string{"Payments"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/NameEnquiryRequest",
						},
						"example": map[string]interface{}{
							"account_number": "1000000001",
							"bank_code":      "044",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Name enquiry successful",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"example": map[string]interface{}{
								"data": map[string]interface{}{
									"account_name": "John Doe",
									"is_internal":  true,
									"currency":     "USD",
								},
								"message": "account name retrieved successfully",
							},
						},
					},
				},
				"400": getErrorResponse("Bad request - invalid parameters"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getExchangeRateEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get exchange rate",
			"description": "Get current exchange rate between two currencies. Rates are locked at transaction initiation time.",
			"operationId": "getExchangeRate",
			"tags":        []string{"Payments"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "from",
					"in":          "query",
					"required":    true,
					"description": "Source currency",
					"schema": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"USD", "EUR", "GBP"},
						"example": "USD",
					},
				},
				{
					"name":        "to",
					"in":          "query",
					"required":    true,
					"description": "Destination currency",
					"schema": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"USD", "EUR", "GBP"},
						"example": "EUR",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Exchange rate retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"example": map[string]interface{}{
								"data": map[string]interface{}{
									"from_currency": "USD",
									"to_currency":   "EUR",
									"rate":          0.85,
								},
								"message": "exchange rate retrieved successfully",
							},
						},
					},
				},
				"400": getErrorResponse("Bad request - invalid currency"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getWalletsEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get user wallets",
			"description": "Get all wallets for the authenticated user with bank account details and cached balances (optimized, not ledger sum).",
			"operationId": "getUserWallets",
			"tags":        []string{"Wallets"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Wallets retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"example": map[string]interface{}{
								"data": []map[string]interface{}{
									{
										"id":             "wallet_user1_usd",
										"currency":       "USD",
										"balance":        100.50,
										"account_number": "1000000001",
										"bank_name":      "Test Bank",
										"bank_code":      "044",
										"account_name":   "John Doe",
										"provider":       "currencycloud",
										"created_at":     "2026-01-11T00:00:00Z",
										"updated_at":     "2026-01-11T00:00:00Z",
									},
								},
								"message": "wallets retrieved successfully",
							},
						},
					},
				},
				"401": getErrorResponse("Unauthorized - missing or invalid X-User-ID"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getCreateInternalTransferEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Create internal transfer",
			"description": "Initiate an internal transfer between users. Uses account number and bank code to identify recipient. Same user transfers (different currency) process immediately. Different user transfers require PIN confirmation.",
			"operationId": "createInternalTransfer",
			"tags":        []string{"Payments"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "Idempotency-Key",
					"in":          "header",
					"required":    true,
					"description": "Unique key for idempotency",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "unique-key-123",
					},
				},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/CreateInternalTransferRequest",
						},
						"example": map[string]interface{}{
							"from_currency":     "USD",
							"to_account_number": "2000000001",
							"to_bank_code":      "044",
							"amount": map[string]interface{}{
								"amount":   100.50,
								"currency": "EUR",
							},
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "Transfer initiated or completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"examples": map[string]interface{}{
								"completed": map[string]interface{}{
									"summary": "Immediate completion (same user)",
									"value": map[string]interface{}{
										"data": map[string]interface{}{
											"id":            "tx-id",
											"status":        "completed",
											"amount":        100.50,
											"currency":      "EUR",
											"exchange_rate": 0.85,
										},
										"message": "transfer completed successfully",
									},
								},
								"initiated": map[string]interface{}{
									"summary": "Requires confirmation (different user)",
									"value": map[string]interface{}{
										"data": map[string]interface{}{
											"id":            "tx-id",
											"status":        "initiated",
											"amount":        100.50,
											"currency":      "EUR",
											"exchange_rate": 0.85,
										},
										"message": "transfer initiated, please confirm with PIN",
									},
								},
							},
						},
					},
				},
				"400": getErrorResponse("Bad request - validation error, insufficient funds, or invalid parameters"),
				"401": getErrorResponse("Unauthorized - missing or invalid X-User-ID"),
				"404": getErrorResponse("Not found - sender wallet or recipient account not found"),
				"409": getErrorResponse("Conflict - duplicate idempotency key"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getCreateExternalTransferEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Create external transfer",
			"description": "Initiate an external transfer to a bank account outside the system. Always requires PIN confirmation. Exchange rate is locked at initiation time.",
			"operationId": "createExternalTransfer",
			"tags":        []string{"Payments"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "Idempotency-Key",
					"in":          "header",
					"required":    true,
					"description": "Unique key for idempotency",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "unique-key-123",
					},
				},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/CreateExternalTransferRequest",
						},
						"example": map[string]interface{}{
							"from_currency":     "USD",
							"to_account_number": "9999999999",
							"to_bank_code":      "044",
							"amount": map[string]interface{}{
								"amount":   50.00,
								"currency": "GBP",
							},
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "External transfer initiated",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"example": map[string]interface{}{
								"data": map[string]interface{}{
									"id":            "tx-id",
									"status":        "initiated",
									"amount":        50.00,
									"currency":      "GBP",
									"exchange_rate": 0.75,
								},
								"message": "external transfer initiated, please confirm with PIN",
							},
						},
					},
				},
				"400": getErrorResponse("Bad request - validation error, insufficient funds, or invalid parameters"),
				"401": getErrorResponse("Unauthorized - missing or invalid X-User-ID"),
				"404": getErrorResponse("Not found - sender wallet not found"),
				"409": getErrorResponse("Conflict - duplicate idempotency key"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getConfirmTransactionEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Confirm transaction",
			"description": "Confirm an initiated transaction with PIN. Transaction must be in 'initiated' status and not expired (10 minutes from creation). Internal transfers complete immediately. External transfers are queued for asynchronous processing.",
			"operationId": "confirmTransaction",
			"tags":        []string{"Payments"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Transaction ID",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "tx-id-123",
					},
				},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ConfirmTransactionRequest",
						},
						"example": map[string]interface{}{
							"pin": "12345",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Transaction confirmed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"examples": map[string]interface{}{
								"internal": map[string]interface{}{
									"summary": "Internal transfer completed",
									"value": map[string]interface{}{
										"data": map[string]interface{}{
											"id":       "tx-id",
											"status":   "completed",
											"amount":   100.50,
											"currency": "EUR",
										},
										"message": "transaction confirmed and completed successfully",
									},
								},
								"external": map[string]interface{}{
									"summary": "External transfer queued",
									"value": map[string]interface{}{
										"data": map[string]interface{}{
											"id":                 "tx-id",
											"status":             "completed",
											"amount":             50.00,
											"currency":           "GBP",
											"provider_reference": "{\"account_number\":\"9999999999\",\"bank_code\":\"044\"}",
										},
										"message": "transaction confirmed and queued for processing",
									},
								},
							},
						},
					},
				},
				"400": getErrorResponse("Bad request - invalid PIN, expired transaction, or insufficient funds"),
				"401": getErrorResponse("Unauthorized - missing or invalid X-User-ID"),
				"404": getErrorResponse("Not found - transaction not found"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getGetTransactionEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get transaction",
			"description": "Get transaction details by ID",
			"operationId": "getTransaction",
			"tags":        []string{"Payments"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Transaction ID",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "tx-id-123",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Transaction retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"example": map[string]interface{}{
								"data": map[string]interface{}{
									"id":              "tx-id",
									"idempotency_key": "key-123",
									"from_wallet_id":  "wallet_user1_usd",
									"to_wallet_id":    "wallet_user2_eur",
									"type":            "internal",
									"amount":          100.50,
									"currency":        "EUR",
									"status":          "completed",
									"exchange_rate":   0.85,
									"created_at":      "2026-01-11T00:00:00Z",
									"updated_at":      "2026-01-11T00:00:00Z",
								},
								"message": "transaction retrieved successfully",
							},
						},
					},
				},
				"401": getErrorResponse("Unauthorized - missing or invalid X-User-ID"),
				"404": getErrorResponse("Not found - transaction not found"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getTransactionHistoryEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get transaction history",
			"description": "Get paginated transaction history using cursor-based pagination. Returns transactions in descending order (newest first).",
			"operationId": "getTransactionHistory",
			"tags":        []string{"Payments"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "cursor",
					"in":          "query",
					"required":    false,
					"description": "Base64-encoded cursor from previous response for pagination",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "base64-encoded-cursor",
					},
				},
				{
					"name":        "limit",
					"in":          "query",
					"required":    false,
					"description": "Number of transactions per page (default: 20, max: 100)",
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 100,
						"default": 20,
						"example": 20,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Transaction history retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"example": map[string]interface{}{
								"data": map[string]interface{}{
									"transactions": []map[string]interface{}{
										{
											"id":         "tx-id-1",
											"type":       "internal",
											"amount":     100.50,
											"currency":   "EUR",
											"status":     "completed",
											"created_at": "2026-01-11T00:00:00Z",
										},
										{
											"id":         "tx-id-2",
											"type":       "external",
											"amount":     50.00,
											"currency":   "GBP",
											"status":     "completed",
											"created_at": "2026-01-10T00:00:00Z",
										},
									},
									"next_cursor": "base64-encoded-cursor",
								},
								"message": "transaction history retrieved successfully",
							},
						},
					},
				},
				"400": getErrorResponse("Bad request - invalid limit parameter"),
				"401": getErrorResponse("Unauthorized - missing or invalid X-User-ID"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getWebhookEndpoint() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Receive webhook",
			"description": "Receive webhook events from payment providers. Events are queued for asynchronous processing.",
			"operationId": "receiveWebhook",
			"tags":        []string{"Webhooks"},
			"security": []map[string]interface{}{
				{"X-User-ID": []string{}},
			},
			"parameters": []map[string]interface{}{
				{
					"name":        "provider",
					"in":          "path",
					"required":    true,
					"description": "Provider name (e.g., currencycloud, dlocal)",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "currencycloud",
					},
				},
				{
					"name":        "reference",
					"in":          "query",
					"required":    false,
					"description": "Provider reference ID (optional, can be in body)",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "TXN-12345",
					},
				},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/WebhookRequest",
						},
						"example": map[string]interface{}{
							"event_type":     "payout.completed",
							"reference":      "TXN-12345",
							"transaction_id": "tx-id",
							"status":         "completed",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Webhook received successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
							"example": map[string]interface{}{
								"data": map[string]interface{}{
									"status": "received",
								},
								"message": "webhook received successfully",
							},
						},
					},
				},
				"400": getErrorResponse("Bad request - invalid webhook payload or missing reference"),
				"401": getErrorResponse("Unauthorized - missing or invalid X-User-ID"),
				"500": getErrorResponse("Internal server error"),
			},
		},
	}
}

func getErrorResponse(description string) map[string]interface{} {
	return map[string]interface{}{
		"description": description,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/ErrorResponse",
				},
				"example": map[string]interface{}{
					"error":   "error_code",
					"message": description,
				},
			},
		},
	}
}

func getSchemas() map[string]interface{} {
	return map[string]interface{}{
		"AmountRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"amount", "currency"},
			"properties": map[string]interface{}{
				"amount": map[string]interface{}{
					"type":        "number",
					"format":      "float",
					"minimum":     0,
					"example":     100.50,
					"description": "Amount in major units (dollars, euros, pounds)",
				},
				"currency": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"USD", "EUR", "GBP"},
					"example": "USD",
				},
			},
		},
		"NameEnquiryRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"account_number", "bank_code"},
			"properties": map[string]interface{}{
				"account_number": map[string]interface{}{
					"type":    "string",
					"example": "1000000001",
				},
				"bank_code": map[string]interface{}{
					"type":    "string",
					"example": "044",
				},
			},
		},
		"CreateInternalTransferRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"from_currency", "to_account_number", "to_bank_code", "amount"},
			"properties": map[string]interface{}{
				"from_currency": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"USD", "EUR", "GBP"},
					"example": "USD",
				},
				"to_account_number": map[string]interface{}{
					"type":    "string",
					"example": "2000000001",
				},
				"to_bank_code": map[string]interface{}{
					"type":    "string",
					"example": "044",
				},
				"amount": map[string]interface{}{
					"$ref": "#/components/schemas/AmountRequest",
				},
			},
		},
		"CreateExternalTransferRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"from_currency", "to_account_number", "to_bank_code", "amount"},
			"properties": map[string]interface{}{
				"from_currency": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"USD", "EUR", "GBP"},
					"example": "USD",
				},
				"to_account_number": map[string]interface{}{
					"type":    "string",
					"example": "9999999999",
				},
				"to_bank_code": map[string]interface{}{
					"type":    "string",
					"example": "044",
				},
				"amount": map[string]interface{}{
					"$ref": "#/components/schemas/AmountRequest",
				},
			},
		},
		"ConfirmTransactionRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"pin"},
			"properties": map[string]interface{}{
				"pin": map[string]interface{}{
					"type":        "string",
					"pattern":     "^[0-9]{5}$",
					"example":     "12345",
					"description": "5-digit numeric PIN",
				},
			},
		},
		"WebhookRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"event_type"},
			"properties": map[string]interface{}{
				"event_type": map[string]interface{}{
					"type":    "string",
					"example": "payout.completed",
				},
				"reference": map[string]interface{}{
					"type":    "string",
					"example": "TXN-12345",
				},
				"transaction_id": map[string]interface{}{
					"type":    "string",
					"example": "tx-id",
				},
				"status": map[string]interface{}{
					"type":    "string",
					"example": "completed",
				},
			},
		},
		"TransactionResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":    "string",
					"example": "tx-id",
				},
				"idempotency_key": map[string]interface{}{
					"type":    "string",
					"example": "key-123",
				},
				"from_wallet_id": map[string]interface{}{
					"type":    "string",
					"example": "wallet_user1_usd",
				},
				"to_wallet_id": map[string]interface{}{
					"type":    "string",
					"example": "wallet_user2_eur",
				},
				"type": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"internal", "external"},
					"example": "internal",
				},
				"amount": map[string]interface{}{
					"type":    "number",
					"format":  "float",
					"example": 100.50,
				},
				"currency": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"USD", "EUR", "GBP"},
					"example": "EUR",
				},
				"status": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"initiated", "pending", "completed", "failed"},
					"example": "completed",
				},
				"provider_name": map[string]interface{}{
					"type":    "string",
					"example": "currencycloud",
				},
				"provider_reference": map[string]interface{}{
					"type":    "string",
					"example": "TXN-12345",
				},
				"exchange_rate": map[string]interface{}{
					"type":    "number",
					"format":  "float",
					"example": 0.85,
				},
				"failure_reason": map[string]interface{}{
					"type":    "string",
					"example": "Insufficient funds",
				},
				"created_at": map[string]interface{}{
					"type":    "string",
					"format":  "date-time",
					"example": "2026-01-11T00:00:00Z",
				},
				"updated_at": map[string]interface{}{
					"type":    "string",
					"format":  "date-time",
					"example": "2026-01-11T00:00:00Z",
				},
			},
		},
		"WalletWithBankAccountResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":    "string",
					"example": "wallet_user1_usd",
				},
				"currency": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"USD", "EUR", "GBP"},
					"example": "USD",
				},
				"balance": map[string]interface{}{
					"type":    "number",
					"format":  "float",
					"example": 100.50,
				},
				"account_number": map[string]interface{}{
					"type":    "string",
					"example": "1000000001",
				},
				"bank_name": map[string]interface{}{
					"type":    "string",
					"example": "Test Bank",
				},
				"bank_code": map[string]interface{}{
					"type":    "string",
					"example": "044",
				},
				"account_name": map[string]interface{}{
					"type":    "string",
					"example": "John Doe",
				},
				"provider": map[string]interface{}{
					"type":    "string",
					"example": "currencycloud",
				},
				"created_at": map[string]interface{}{
					"type":    "string",
					"format":  "date-time",
					"example": "2026-01-11T00:00:00Z",
				},
				"updated_at": map[string]interface{}{
					"type":    "string",
					"format":  "date-time",
					"example": "2026-01-11T00:00:00Z",
				},
			},
		},
		"TestUserDataResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"user_id": map[string]interface{}{
					"type":    "string",
					"example": "user_1",
				},
				"name": map[string]interface{}{
					"type":    "string",
					"example": "John Doe",
				},
				"pin": map[string]interface{}{
					"type":    "string",
					"example": "12345",
				},
				"wallets": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"$ref": "#/components/schemas/WalletWithBankAccountResponse",
					},
				},
			},
		},
		"TestUsersResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"users": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"$ref": "#/components/schemas/TestUserDataResponse",
					},
				},
			},
		},
		"SuccessResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{
					"description": "Response data (structure varies by endpoint)",
				},
				"message": map[string]interface{}{
					"type":    "string",
					"example": "Operation completed successfully",
				},
			},
		},
		"ErrorResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "string",
					"example": "error_code",
				},
				"message": map[string]interface{}{
					"type":    "string",
					"example": "Human-readable error message",
				},
			},
		},
	}
}
