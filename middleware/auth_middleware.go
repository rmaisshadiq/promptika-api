package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmaisshadiq/critical-prompt-api/services"
	"github.com/rmaisshadiq/critical-prompt-api/utils"
)

// JWTAuth is a Gin middleware that validates the Authorization header
// and injects the authenticated user's ID into the context.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		// Expect format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header must be in the format: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := services.ValidateToken(parts[1])
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// Store user ID in context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
