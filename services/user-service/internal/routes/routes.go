package routes

import (
	"github.com/deepwrite/user-service/internal/handler"
	"github.com/deepwrite/user-service/internal/middleware"
	"github.com/deepwrite/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

func Setup(
	router *gin.Engine,
	userHandler *handler.UserHandler,
	teamHandler *handler.TeamHandler,
	userService *service.UserService,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)

			auth := users.Group("", middleware.JWTAuth(userService))
			{
				auth.GET("/me", userHandler.GetMe)
				auth.PUT("/me", userHandler.UpdateMe)
			}
		}

		teams := api.Group("/teams", middleware.JWTAuth(userService))
		{
			teams.POST("", teamHandler.Create)
			teams.GET("", teamHandler.List)
			teams.GET("/:id", teamHandler.Get)
			teams.PUT("/:id", teamHandler.Update)
			teams.DELETE("/:id", teamHandler.Delete)

			teams.POST("/:id/members", teamHandler.AddMember)
			teams.GET("/:id/members", teamHandler.GetMembers)
			teams.PUT("/:id/members/:user_id", teamHandler.UpdateMember)
			teams.DELETE("/:id/members/:user_id", teamHandler.RemoveMember)
		}
	}
}
