package handlers

import (
	"strconv"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/handlers/requests"
	"github.com/IfedayoAwe/payment-processing-service/middleware"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	service "github.com/IfedayoAwe/payment-processing-service/services"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/labstack/echo/v4"
)

type PaymentHandler interface {
	CreateInternalTransfer(c echo.Context) error
	CreateExternalTransfer(c echo.Context) error
	ConfirmTransaction(c echo.Context) error
	GetTransaction(c echo.Context) error
	GetTransactionHistory(c echo.Context) error
	GetExchangeRate(c echo.Context) error
	GetUserWallets(c echo.Context) error
	GetTestUsers(c echo.Context) error
}

type paymentHandler struct {
	paymentService service.PaymentService
	walletService  service.WalletService
	queries        *gen.Queries
}

func newPaymentHandler(paymentService service.PaymentService, walletService service.WalletService, queries *gen.Queries) PaymentHandler {
	return &paymentHandler{
		paymentService: paymentService,
		walletService:  walletService,
		queries:        queries,
	}
}

func (ph *paymentHandler) CreateInternalTransfer(c echo.Context) error {
	var req requests.CreateInternalTransferRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ValidationError(c, utils.FormatValidationErrors(err))
	}

	fromUserID := middleware.GetUserID(c)

	fromCurrency, err := money.ParseCurrency(req.FromCurrency)
	if err != nil {
		return utils.BadRequest(c, "invalid from currency")
	}

	toAmount, err := req.Amount.ToMoney()
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return utils.BadRequest(c, "Idempotency-Key header is required")
	}

	transaction, err := ph.paymentService.CreateInternalTransfer(c.Request().Context(), fromUserID, req.ToAccountNumber, req.ToBankCode, fromCurrency, toAmount, idempotencyKey)
	if err != nil {
		return utils.HandleError(c, err)
	}

	response := models.TransactionToResponse(transaction)
	if transaction.Status == models.TransactionStatusCompleted {
		return utils.Created(c, response, "transfer completed successfully")
	}

	return utils.Created(c, response, "transfer initiated, please confirm with PIN")
}

func (ph *paymentHandler) ConfirmTransaction(c echo.Context) error {
	transactionID := c.Param("id")
	if transactionID == "" {
		return utils.BadRequest(c, "transaction ID is required")
	}

	var req requests.ConfirmTransactionRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := req.Validate(); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return utils.ValidationError(c, utils.FormatValidationErrors(err))
	}

	userID := middleware.GetUserID(c)

	transaction, err := ph.paymentService.ConfirmTransaction(c.Request().Context(), transactionID, userID, req.PIN)
	if err != nil {
		return utils.HandleError(c, err)
	}

	response := models.TransactionToResponse(transaction)
	if transaction.Status == models.TransactionStatusCompleted {
		return utils.Success(c, response, "transaction confirmed and completed successfully")
	}

	return utils.Success(c, response, "transaction confirmed and queued for processing")
}

func (ph *paymentHandler) GetTransaction(c echo.Context) error {
	transactionID := c.Param("id")
	if transactionID == "" {
		return utils.BadRequest(c, "transaction ID is required")
	}

	transaction, err := ph.paymentService.GetTransactionByID(c.Request().Context(), transactionID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	response := models.TransactionToResponse(transaction)
	return utils.Success(c, response, "transaction retrieved successfully")
}

func (ph *paymentHandler) CreateExternalTransfer(c echo.Context) error {
	var req requests.CreateExternalTransferRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ValidationError(c, utils.FormatValidationErrors(err))
	}

	userID := middleware.GetUserID(c)

	fromCurrency, err := money.ParseCurrency(req.FromCurrency)
	if err != nil {
		return utils.BadRequest(c, "invalid from currency")
	}

	toAmount, err := req.Amount.ToMoney()
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return utils.BadRequest(c, "Idempotency-Key header is required")
	}

	transaction, err := ph.paymentService.CreateExternalTransfer(c.Request().Context(), userID, req.ToAccountNumber, req.ToBankCode, fromCurrency, toAmount, idempotencyKey)
	if err != nil {
		return utils.HandleError(c, err)
	}

	response := models.TransactionToResponse(transaction)
	return utils.Created(c, response, "external transfer initiated, please confirm with PIN")
}

func (ph *paymentHandler) GetExchangeRate(c echo.Context) error {
	fromCurrencyStr := c.QueryParam("from")
	toCurrencyStr := c.QueryParam("to")

	if fromCurrencyStr == "" {
		return utils.BadRequest(c, "from currency parameter is required")
	}

	if toCurrencyStr == "" {
		return utils.BadRequest(c, "to currency parameter is required")
	}

	fromCurrency, err := money.ParseCurrency(fromCurrencyStr)
	if err != nil {
		return utils.BadRequest(c, "invalid from currency")
	}

	toCurrency, err := money.ParseCurrency(toCurrencyStr)
	if err != nil {
		return utils.BadRequest(c, "invalid to currency")
	}

	rate, err := ph.paymentService.GetExchangeRate(c.Request().Context(), fromCurrency, toCurrency)
	if err != nil {
		return utils.HandleError(c, err)
	}

	return utils.Success(c, map[string]interface{}{
		"from_currency": fromCurrency.String(),
		"to_currency":   toCurrency.String(),
		"rate":          rate,
	}, "exchange rate retrieved successfully")
}

func (ph *paymentHandler) GetUserWallets(c echo.Context) error {
	userID := middleware.GetUserID(c)

	wallets, err := ph.walletService.GetUserWallets(c.Request().Context(), userID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	response := make([]*models.WalletWithBankAccountResponse, len(wallets))
	for i, wallet := range wallets {
		response[i] = models.WalletWithBankAccountToResponse(wallet)
	}

	return utils.Success(c, response, "wallets retrieved successfully")
}

func (ph *paymentHandler) GetTestUsers(c echo.Context) error {
	ctx := c.Request().Context()

	testUserIDs := []string{"user_1", "user_2"}
	testPIN := "12345"

	var usersResponse []*models.TestUserDataResponse

	for _, userID := range testUserIDs {
		user, err := ph.queries.GetUserByID(ctx, userID)
		if err != nil {
			return utils.HandleError(c, err)
		}

		wallets, err := ph.walletService.GetUserWallets(ctx, userID)
		if err != nil {
			return utils.HandleError(c, err)
		}

		var name *string
		if user.Name.Valid {
			name = &user.Name.String
		}

		walletsResponse := make([]*models.WalletWithBankAccountResponse, len(wallets))
		for i, wallet := range wallets {
			walletsResponse[i] = models.WalletWithBankAccountToResponse(wallet)
		}

		testUser := &models.TestUserDataResponse{
			UserID:  user.UserID,
			Name:    name,
			PIN:     testPIN,
			Wallets: walletsResponse,
		}

		usersResponse = append(usersResponse, testUser)
	}

	response := &models.TestUsersResponseDTO{
		Users: usersResponse,
	}

	return utils.Success(c, response, "test users retrieved successfully")
}

func (ph *paymentHandler) GetTransactionHistory(c echo.Context) error {
	userID := middleware.GetUserID(c)

	cursor := c.QueryParam("cursor")
	limitStr := c.QueryParam("limit")

	limit := int32(20)
	if limitStr != "" {
		parsed, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil || parsed <= 0 {
			return utils.BadRequest(c, "invalid limit parameter")
		}
		limit = int32(parsed)
	}

	history, err := ph.paymentService.GetTransactionHistory(c.Request().Context(), userID, cursor, limit)
	if err != nil {
		return utils.HandleError(c, err)
	}

	transactionsResponse := make([]*models.TransactionResponse, len(history.Transactions))
	for i, tx := range history.Transactions {
		transactionsResponse[i] = models.TransactionToResponse(tx)
	}

	response := &models.TransactionHistoryResponseDTO{
		Transactions: transactionsResponse,
		NextCursor:   history.NextCursor,
	}

	return utils.Success(c, response, "transaction history retrieved successfully")
}
