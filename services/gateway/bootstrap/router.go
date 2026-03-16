package bootstrap

import (
	"time"

	gatewaylogger "github.com/deepwrite/serivces/gateway/pkg/logger"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/routers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(r *gin.Engine) {
	registerGlobalMiddleware(r)

	routers.SetupAPIRoutes(r)

	registerSwagger(r)

	registerNotFoundHandler(r)
}

func registerGlobalMiddleware(r *gin.Engine) {
	r.Use(
		gatewaylogger.GinRecovery(),
		gatewaylogger.GinLogger(),
		cors.New(cors.Config{
			AllowOrigins: []string{
				"http://localhost:3000",
				"http://127.0.0.1:3000",
				"http://localhost:3001",
				"http://127.0.0.1:3001",
			},
			AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders: []string{"Content-Length"},
			MaxAge:        12 * time.Hour,
		}),
	)
}

func registerNotFoundHandler(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		response.Failed(c, 404, "Route not found")
	})
}

func registerSwagger(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
