package requests

type WebhookRequest struct {
	EventType     string  `json:"event_type" validate:"required"`
	Reference     string  `json:"reference"`
	TransactionID *string `json:"transaction_id"`
	Status        string  `json:"status"`
}
