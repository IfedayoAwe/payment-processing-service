package handlers

import (
	service "github.com/IfedayoAwe/payment-processing-service/services"
)

type Handlers struct{}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{}
}
