package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type SecurityConfig struct {
	HTTPS            bool
	TurnstileEnabled bool
}

func SecurityHeaders(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.HTTPS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		scriptSrc := "'self' 'unsafe-inline' https://cdn.jsdelivr.net"
		frameSrc := "'none'"
		connectSrc := "'self'"
		if config.TurnstileEnabled {
			scriptSrc += " https://challenges.cloudflare.com"
			frameSrc = "https://challenges.cloudflare.com"
			connectSrc += " https://challenges.cloudflare.com"
		}

		csp := []string{
			"default-src 'self'",
			"script-src " + scriptSrc,
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net",
			"font-src 'self' https://cdn.jsdelivr.net",
			"img-src 'self' data:",
			"connect-src " + connectSrc,
			"frame-src " + frameSrc,
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
