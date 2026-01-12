package services

import (
	"strconv"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
)

func mapTransaction(t gen.Transaction) *models.Transaction {
	var fromWalletID *string
	if t.FromWalletID.Valid {
		fromWalletID = &t.FromWalletID.String
	}

	var toWalletID *string
	if t.ToWalletID.Valid {
		toWalletID = &t.ToWalletID.String
	}

	var traceID, providerName, providerRef, failureReason *string
	var exchangeRate *float64
	if t.TraceID.Valid {
		traceID = &t.TraceID.String
	}
	if t.ProviderName.Valid {
		providerName = &t.ProviderName.String
	}
	if t.ProviderReference.Valid {
		providerRef = &t.ProviderReference.String
	}
	if t.FailureReason.Valid {
		failureReason = &t.FailureReason.String
	}
	if t.ExchangeRate.Valid {
		if f, err := strconv.ParseFloat(t.ExchangeRate.String, 64); err == nil && f > 0 {
			exchangeRate = &f
		}
	}

	return &models.Transaction{
		ID:                t.ID,
		IdempotencyKey:    t.IdempotencyKey,
		TraceID:           traceID,
		FromWalletID:      fromWalletID,
		ToWalletID:        toWalletID,
		Type:              models.TransactionType(t.Type),
		Amount:            t.Amount,
		Currency:          t.Currency,
		Status:            models.TransactionStatus(t.Status),
		ProviderName:      providerName,
		ProviderReference: providerRef,
		ExchangeRate:      exchangeRate,
		FailureReason:     failureReason,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
}
