package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func TraceIDHeader() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqCtx := c.UserContext()
		traceID := trace.SpanFromContext(reqCtx).SpanContext().TraceID().String()
		c.Set(RequestIDHeader, traceID)
		c.Locals("requestID", traceID)
		return c.Next()
	}
}
