package middleware

import (
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/labstack/echo/v4"
)

func TraceIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			traceID := c.Request().Header.Get(utils.TraceIDHeader)
			if traceID == "" {
				traceID = utils.GenerateTraceID()
			}

			ctx = utils.WithTraceID(ctx, traceID)
			c.SetRequest(c.Request().WithContext(ctx))

			c.Response().Header().Set(utils.TraceIDHeader, traceID)

			return next(c)
		}
	}
}
