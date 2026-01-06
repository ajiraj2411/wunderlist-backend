package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf(
			"%s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			status,
			latency,
		)
	}
}
