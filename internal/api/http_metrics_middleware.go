package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"job4j.ru/share_trip/internal/observability/metrics"
)

func NewHTTPMetricsMiddleware(m *metrics.Metrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		started := time.Now()

		err := c.Next()

		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}
		// Prometheus scrapes this path; do not count it as application traffic.
		if path == MetricsInfo {
			return err
		}

		status := strconv.Itoa(c.Response().StatusCode())

		m.HTTPRequestTotal.WithLabelValues(
			c.Method(),
			path,
			status,
		).Inc()

		m.HTTPRequestDuration.WithLabelValues(
			c.Method(),
			path,
			status,
		).Observe(time.Since(started).Seconds())

		return err
	}
}
