package routers

import (
	"github.com/deepwrite/serivces/gateway/handler"
	"github.com/deepwrite/serivces/gateway/middleware"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

func SetupAPIRoutes(r *gin.Engine) {

	v1 := r.Group("/api/v1")
	{
		v1.GET("/", func(c *gin.Context) {
			response.Success(c, response.SuccessCode, gin.H{"message": "Welcome to DeepWrite API"})
		})

		// Auth routes
		authGroup := v1.Group("/auth")
		authHandler := new(handler.AuthHandler)
		userHandler := new(handler.UserHandler)
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)

			protected := authGroup.Group("")
			protected.Use(middleware.AuthMiddleware())
			protected.POST("/logout", authHandler.Logout)
			protected.GET("/me", authHandler.Me)
		}

		userGroup := v1.Group("/user")
		{
			userGroup.POST("/forgot-password", userHandler.ForgotPassword)
			userGroup.POST("/reset-password", userHandler.ResetPassword)

			protected := userGroup.Group("")
			protected.Use(middleware.AuthMiddleware())
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)
			protected.POST("/avatar", userHandler.UploadAvatar)
			protected.POST("/change-password", userHandler.ChangePassword)
			protected.POST("/change-email", userHandler.ChangeEmail)
		}

		workspaceGroup := v1.Group("/workspaces")
		workspaceGroup.Use(middleware.AuthMiddleware())
		workspaceHandler := new(handler.WorkspaceHandler)
		{
			workspaceGroup.GET("", workspaceHandler.List)
			workspaceGroup.POST("", workspaceHandler.Create)
			workspaceGroup.GET("/:id", workspaceHandler.GetByID)
			workspaceGroup.PUT("/:id", workspaceHandler.Update)
			workspaceGroup.DELETE("/:id", workspaceHandler.Delete)

			workspaceGroup.GET("/:id/folders", workspaceHandler.ListFolders)
			workspaceGroup.POST("/:id/folders", workspaceHandler.CreateFolder)
			workspaceGroup.GET("/:id/folders/:folder_id", workspaceHandler.GetFolderByID)
			workspaceGroup.PUT("/:id/folders/:folder_id", workspaceHandler.UpdateFolder)
			workspaceGroup.DELETE("/:id/folders/:folder_id", workspaceHandler.DeleteFolder)
			workspaceGroup.GET("/:id/files", workspaceHandler.ListFiles)
			workspaceGroup.GET("/:id/files/:file_id", workspaceHandler.GetFileByID)
			workspaceGroup.POST("/:id/files", workspaceHandler.UploadFile)
			workspaceGroup.POST("/:id/files/complete", workspaceHandler.CompleteUpload)
			workspaceGroup.POST("/:id/files/batch-delete", workspaceHandler.BatchDeleteFiles)
			workspaceGroup.DELETE("/:id/files/:file_id", workspaceHandler.DeleteFile)

			workspaceGroup.GET("/:id/invitations", workspaceHandler.ListInvitations)
			workspaceGroup.POST("/:id/invitations", workspaceHandler.CreateInvitation)
			workspaceGroup.POST("/:id/invitations/:invite_id/revoke", workspaceHandler.RevokeInvitation)
		}

		invitationGroup := v1.Group("/invitations")
		invitationGroup.Use(middleware.AuthMiddleware())
		{
			invitationGroup.GET("/me", workspaceHandler.ListMyInvitations)
			invitationGroup.POST("/:invite_id/accept", workspaceHandler.AcceptInvitation)
			invitationGroup.POST("/:invite_id/reject", workspaceHandler.RejectInvitation)
		}
	}
}
