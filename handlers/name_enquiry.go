package handlers

import (
	"github.com/IfedayoAwe/payment-processing-service/handlers/requests"
	"github.com/IfedayoAwe/payment-processing-service/models"
	service "github.com/IfedayoAwe/payment-processing-service/services"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/labstack/echo/v4"
)

type NameEnquiryHandler interface {
	EnquireAccountName(c echo.Context) error
}

type nameEnquiryHandler struct {
	nameEnquiryService service.NameEnquiryService
}

func newNameEnquiryHandler(nameEnquiryService service.NameEnquiryService) NameEnquiryHandler {
	return &nameEnquiryHandler{
		nameEnquiryService: nameEnquiryService,
	}
}

func (neh *nameEnquiryHandler) EnquireAccountName(c echo.Context) error {
	var req requests.NameEnquiryRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ValidationError(c, utils.FormatValidationErrors(err))
	}

	result, err := neh.nameEnquiryService.EnquireAccountName(c.Request().Context(), req.AccountNumber, req.BankCode)
	if err != nil {
		return utils.HandleError(c, err)
	}

	return utils.Success(c, models.NameEnquiryResponse{
		AccountName: result.AccountName,
		IsInternal:  result.IsInternal,
		Currency:    result.Currency.String(),
	}, "account name retrieved successfully")
}
