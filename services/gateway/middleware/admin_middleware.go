package middleware

import (
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/jwt"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetString("user_id"))
		if userID == "" {
			response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
			c.Abort()
			return
		}

		userRole := strings.TrimSpace(c.GetString("user_role"))
		if !strings.EqualFold(userRole, "admin") {
			response.Failed(c, response.ErrorForbiddenCode, "需要管理员权限")
			c.Abort()
			return
		}

		currentClaims, ok := jwt.GetCurrentClaims(c)
		if !ok || currentClaims == nil {
			response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
			c.Abort()
			return
		}
		if !strings.EqualFold(strings.TrimSpace(currentClaims.Scope), "admin") {
			response.Failed(c, response.ErrorForbiddenCode, "需要管理员专用令牌")
			c.Abort()
			return
		}

		c.Next()
	}
}
