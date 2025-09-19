package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/pkg/logger"
)

// LoggingMiddleware creates a Gin middleware for structured logging
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Use structured logger instead of default Gin logger
		log.LogRequest(
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ClientIP,
			param.StatusCode,
			param.Latency.String(),
		)
		return "" // Return empty string to prevent default logging
	})
}

// ErrorLoggingMiddleware logs errors that occur during request processing
func ErrorLoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log any errors that occurred during request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.LogError(err.Err, "Request processing error",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"status_code", c.Writer.Status(),
					"client_ip", c.ClientIP(),
				)
			}
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Add request ID to logger context
		contextLogger := log.WithField("request_id", requestID)
		c.Set("logger", contextLogger)

		c.Next()
	}
}

// generateRequestID creates a simple request ID (you might want to use UUID in production)
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
