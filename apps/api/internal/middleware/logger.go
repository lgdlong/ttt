package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequestLogger logs incoming HTTP requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build log message
		logEvent := log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", raw).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("ip", c.ClientIP())

		// Log errors if present
		if len(c.Errors) > 0 {
			logEvent.Str("errors", c.Errors.String())
		}

		logEvent.Msg("HTTP Request")
	}
}
