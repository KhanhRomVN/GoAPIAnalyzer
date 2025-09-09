package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"goapianalyzer/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// RequestLoggerMiddleware logs incoming HTTP requests and responses
func RequestLoggerMiddleware() gin.HandlerFunc {
	log := logger.GetLogger()

	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Capture request start time
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response body writer
		w := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = w

		// Log request
		log.WithFields(map[string]interface{}{
			"request_id":     requestID,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"query":          c.Request.URL.RawQuery,
			"user_agent":     c.Request.UserAgent(),
			"remote_addr":    c.ClientIP(),
			"content_type":   c.ContentType(),
			"content_length": c.Request.ContentLength,
			"request_body":   string(requestBody),
		}).Info("HTTP Request")

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get response body
		responseBody := w.body.String()

		// Log response
		logFields := map[string]interface{}{
			"request_id":    requestID,
			"status_code":   c.Writer.Status(),
			"duration_ms":   duration.Milliseconds(),
			"response_size": c.Writer.Size(),
			"response_body": responseBody,
		}

		if len(c.Errors) > 0 {
			logFields["errors"] = c.Errors.String()
			log.WithFields(logFields).Error("HTTP Response with errors")
		} else {
			switch {
			case c.Writer.Status() >= 500:
				log.WithFields(logFields).Error("HTTP Response - Server Error")
			case c.Writer.Status() >= 400:
				log.WithFields(logFields).Warn("HTTP Response - Client Error")
			case c.Writer.Status() >= 300:
				log.WithFields(logFields).Info("HTTP Response - Redirect")
			default:
				log.WithFields(logFields).Info("HTTP Response - Success")
			}
		}
	}
}

// SimpleLoggerMiddleware provides basic request logging
func SimpleLoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
		Output:    gin.DefaultWriter,
		SkipPaths: []string{"/health", "/ping"},
	})
}

// ErrorLoggerMiddleware logs errors that occur during request processing
func ErrorLoggerMiddleware() gin.HandlerFunc {
	log := logger.GetLogger()

	return func(c *gin.Context) {
		c.Next()

		// Log any errors that occurred
		for _, err := range c.Errors {
			log.WithFields(map[string]interface{}{
				"request_id": c.GetString("request_id"),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"error_type": err.Type,
				"error":      err.Error(),
			}).Error("Request processing error")
		}
	}
}

// RecoveryLoggerMiddleware logs panics and recovers from them
func RecoveryLoggerMiddleware() gin.HandlerFunc {
	log := logger.GetLogger()

	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.WithFields(map[string]interface{}{
			"request_id": c.GetString("request_id"),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"panic":      recovered,
		}).Error("Request panic recovered")

		c.AbortWithStatus(500)
	})
}
