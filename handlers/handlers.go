package handlers

import (
	service "github.com/IfedayoAwe/payment-processing-service/services"
)

type Handlers struct {
	services *service.Services
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		services: services,
	}
}
