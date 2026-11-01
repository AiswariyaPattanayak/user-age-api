package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

// RequestID reuses an incoming X-Request-ID if present, otherwise generates one,
// and injects it into the response headers.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Locals("requestId", id)
		c.Set(RequestIDHeader, id)
		return c.Next()
	}
}

// RequestLogger logs method, path, status, and duration for every request.
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		// Run the app's error handler here so the logged status is the real one.
		if err != nil {
			if herr := c.App().ErrorHandler(c, err); herr != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
			err = nil
		}
		status := c.Response().StatusCode()

		log.Info("request completed",
			zap.Any("request_id", c.Locals("requestId")),
			zap.String("method", c.Method()),
			zap.String("path", c.OriginalURL()),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
		)
		return err
	}
}
