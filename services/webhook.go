package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type WebhookService interface {
	ProcessWebhookEvent(ctx context.Context, providerName, eventType, providerReference string, transactionID *string, payload []byte) error
}

type webhookService struct {
	queries *gen.Queries
}

func (s *Services) Webhook() WebhookService {
	return &webhookService{
		queries: s.queries,
	}
}

func (ws *webhookService) ProcessWebhookEvent(ctx context.Context, providerName, eventType, providerReference string, transactionID *string, payload []byte) error {
	var txID sql.NullString
	if transactionID != nil {
		txID = sql.NullString{String: *transactionID, Valid: true}
	}

	_, err := ws.queries.CreateWebhookEvent(ctx, gen.CreateWebhookEventParams{
		ProviderName:      providerName,
		EventType:         eventType,
		ProviderReference: providerReference,
		TransactionID:     txID,
		Payload:           json.RawMessage(payload),
	})
	if err != nil {
		return utils.ServerErr(fmt.Errorf("create webhook event: %w", err))
	}

	return nil
}
