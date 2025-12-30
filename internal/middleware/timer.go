package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func RequestTimer() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("startTime", time.Now())
		c.Next()
	}
}
