package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type SecurityConfig struct {
	HTTPS bool
}
func SecurityHeaders(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.HTTPS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net",
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net",
			"font-src 'self' https://cdn.jsdelivr.net",
			"img-src 'self' data:",
			"connect-src 'self'",
			"frame-ancestors 'none'",
		}
		c.Header("Content-Security-Policy", strings.Join(csp, "; "))
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}

func ForceHTTPS(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.HTTPS && c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {

			host := c.Request.Host
			target := "https://" + host + c.Request.URL.Path
			if len(c.Request.URL.RawQuery) > 0 {
				target += "?" + c.Request.URL.RawQuery
			}
			c.Redirect(301, target)
			c.Abort()
			return
		}
		c.Next()
	}
}
