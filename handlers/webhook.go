package handlers

import (
	"encoding/json"
	"io"

	"github.com/IfedayoAwe/payment-processing-service/handlers/requests"
	"github.com/IfedayoAwe/payment-processing-service/queue"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/labstack/echo/v4"
)

type WebhookHandler interface {
	ReceiveWebhook(c echo.Context) error
}

type webhookHandler struct {
	queue queue.Queue
}

func newWebhookHandler(queue queue.Queue) WebhookHandler {
	return &webhookHandler{
		queue: queue,
	}
}

func (wh *webhookHandler) ReceiveWebhook(c echo.Context) error {
	providerName := c.Param("provider")
	if providerName == "" {
		return utils.BadRequest(c, "provider name is required")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	var webhookReq requests.WebhookRequest
	if err := json.Unmarshal(body, &webhookReq); err != nil {
		return utils.BadRequest(c, "invalid webhook payload")
	}

	if err := c.Validate(&webhookReq); err != nil {
		return utils.ValidationError(c, utils.FormatValidationErrors(err))
	}

	providerRef := c.QueryParam("reference")
	if providerRef == "" {
		providerRef = webhookReq.Reference
	}
	if providerRef == "" {
		return utils.BadRequest(c, "provider reference is required")
	}

	payload := queue.WebhookJobPayload{
		ProviderName:      providerName,
		EventType:         webhookReq.EventType,
		ProviderReference: providerRef,
		TransactionID:     webhookReq.TransactionID,
		Payload:           body,
	}

	if err := wh.queue.Enqueue(c.Request().Context(), queue.JobTypeWebhook, payload); err != nil {
		return utils.HandleError(c, err)
	}

	return utils.Success(c, map[string]string{"status": "received"}, "webhook received successfully")
}
