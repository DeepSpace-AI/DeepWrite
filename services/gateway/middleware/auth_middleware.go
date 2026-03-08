package middleware

import (
	"github.com/deepwrite/serivces/gateway/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return jwt.AuthMiddleware()
}
