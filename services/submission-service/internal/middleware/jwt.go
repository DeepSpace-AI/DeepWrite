package middleware

import (
	"net/http"
	"strings"

	"github.com/deepwrite/submission-service/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTAuth struct {
	cfg *config.JWTConfig
}

func NewJWTAuth(cfg *config.JWTConfig) *JWTAuth {
	return &JWTAuth{cfg: cfg}
}

func (a *JWTAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Missing authorization header"}})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid authorization format"}})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(a.cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid token"}})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid token claims"}})
			c.Abort()
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid user ID in token"}})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
