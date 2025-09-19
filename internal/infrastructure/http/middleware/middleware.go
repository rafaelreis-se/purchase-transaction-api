package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure appropriately for production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already set
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// ErrorHandler middleware for consistent error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Log the error using structured logging
			requestID, _ := c.Get("request_id")
			slog.Error("Request error",
				"request_id", requestID,
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)

			// Return error response if not already sent
			if !c.Writer.Written() {
				c.JSON(500, gin.H{
					"error":      "Internal server error",
					"request_id": requestID,
				})
			}
		}
	}
}

// Logger middleware for request logging (alternative to gin.Logger())
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		requestID := param.Keys["request_id"]
		if requestID == nil {
			requestID = "unknown"
		}

		return fmt.Sprintf("[%s] %s - [%s] \"%s %s %s\" %d %s \"%s\" \"%s\" (%d)\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.ClientIP,
			requestID,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
			param.BodySize,
		)
	})
}
