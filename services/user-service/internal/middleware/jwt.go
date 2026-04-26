package middleware

import (
	"net/http"
	"strings"

	"github.com/deepwrite/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

func JWTAuth(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Missing authorization header"}})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid authorization header format"}})
			c.Abort()
			return
		}

		claims, err := userService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid or expired token"}})
			c.Abort()
			return
		}

		c.Set("userID", (*claims)["sub"])
		c.Set("email", (*claims)["email"])
		c.Set("tier", (*claims)["tier"])
		c.Next()
	}
}
