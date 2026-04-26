package routes

import (
	"github.com/deepwrite/project-service/internal/handler"
	"github.com/deepwrite/project-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, projectHandler *handler.ProjectHandler, jwtAuth *middleware.JWTAuth) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	auth := jwtAuth.Handler()

	projects := api.Group("/projects", auth)
	{
		projects.POST("", projectHandler.Create)
		projects.GET("", projectHandler.List)
		projects.GET("/:id", projectHandler.Get)
		projects.PUT("/:id", projectHandler.Update)
		projects.DELETE("/:id", projectHandler.Delete)
		projects.POST("/:id/archive", projectHandler.Archive)
		projects.POST("/:id/unarchive", projectHandler.Unarchive)

		projects.POST("/:id/members", projectHandler.AddMember)
		projects.GET("/:id/members", projectHandler.GetMembers)
		projects.PUT("/:id/members/:user_id", projectHandler.UpdateMember)
		projects.DELETE("/:id/members/:user_id", projectHandler.RemoveMember)
	}
}
