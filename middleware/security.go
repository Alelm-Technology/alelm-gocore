package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeaders(options ...SecurityOptions) gin.HandlerFunc {
	opts := defaultSecurityOptions()
	if len(options) > 0 {
		opts = options[0]
	}
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "0")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()")

		if opts.HSTS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if opts.CSP != "" {
			c.Header("Content-Security-Policy", opts.CSP)
		}
		c.Next()
	}
}

type SecurityOptions struct {
	HSTS bool
	CSP  string
}

func defaultSecurityOptions() SecurityOptions {
	return SecurityOptions{
		HSTS: true,
		CSP:  "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'",
	}
}
