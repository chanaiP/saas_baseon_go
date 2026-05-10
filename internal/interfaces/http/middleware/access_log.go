package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		requestID, _ := c.Get(RequestIDKey)
		userID, _ := c.Get("user_id")
		tenantID, _ := c.Get("tenant_id")
		latency := time.Since(start)
		entry := map[string]interface{}{
			"level":       "info",
			"event":       "http_request",
			"request_id":  requestID,
			"tenant_id":   tenantID,
			"user_id":     userID,
			"method":      c.Request.Method,
			"path":        c.FullPath(),
			"status":      c.Writer.Status(),
			"latency_ms":  latency.Milliseconds(),
			"slow":        latency >= time.Second,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"error_count": len(c.Errors),
		}
		raw, err := json.Marshal(entry)
		if err != nil {
			log.Printf(`{"level":"error","event":"http_request_log_failed"}`)
			return
		}
		log.Print(string(raw))
	}
}
