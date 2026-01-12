package handlers

import (
	service "github.com/IfedayoAwe/payment-processing-service/services"
)

type Handlers struct {
	Payment     PaymentHandler
	NameEnquiry NameEnquiryHandler
	Webhook     WebhookHandler
}

func NewHandlers(services *service.Services) *Handlers {
	paymentHandler := newPaymentHandler(services.Payment, services.Wallet, services.Queries)
	nameEnquiryHandler := newNameEnquiryHandler(services.NameEnquiry)
	webhookHandler := newWebhookHandler(services.Queue)

	return &Handlers{
		Payment:     paymentHandler,
		NameEnquiry: nameEnquiryHandler,
		Webhook:     webhookHandler,
	}
}
