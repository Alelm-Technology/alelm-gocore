package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

	isAllowed := len(allowedOrigins) == 0
	if !isAllowed {
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				isAllowed = true
				break
			}
		}
	}
	if !isAllowed && origin != "" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
	} else if len(allowedOrigins) == 0 {
		c.Header("Access-Control-Allow-Origin", "*")
	}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
