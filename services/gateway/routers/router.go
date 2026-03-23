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
			authGroup.POST("/admin/login", authHandler.AdminLogin)
			authGroup.POST("/refresh", authHandler.RefreshToken)

			protected := authGroup.Group("")
			protected.Use(middleware.AuthMiddleware())
			protected.POST("/logout", authHandler.Logout)
			protected.GET("/me", authHandler.Me)
		}

		userGroup := v1.Group("/user")
		workspaceHandler := new(handler.WorkspaceHandler)
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

		dashboardGroup := v1.Group("/dashboard")
		dashboardGroup.Use(middleware.AuthMiddleware())
		{
			dashboardGroup.GET("/overview", workspaceHandler.DashboardOverview)
		}

		workspaceGroup := v1.Group("/workspaces")
		workspaceGroup.Use(middleware.AuthMiddleware())
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
			workspaceGroup.PUT(":id/files/:file_id", workspaceHandler.UpdateFile)
			workspaceGroup.DELETE("/:id/files/:file_id", workspaceHandler.DeleteFile)

			workspaceGroup.GET("/:id/files/:file_id/annotations", workspaceHandler.ListAnnotations)
			workspaceGroup.POST("/:id/files/:file_id/annotations", workspaceHandler.CreateAnnotation)
			workspaceGroup.PUT("/:id/files/:file_id/annotations/:annotation_id", workspaceHandler.UpdateAnnotation)
			workspaceGroup.DELETE("/:id/files/:file_id/annotations/:annotation_id", workspaceHandler.DeleteAnnotation)

			workspaceGroup.GET("/:id/invitations", workspaceHandler.ListInvitations)
			workspaceGroup.POST("/:id/invitations", workspaceHandler.CreateInvitation)
			workspaceGroup.POST("/:id/invitations/:invite_id/revoke", workspaceHandler.RevokeInvitation)
			workspaceGroup.GET("/:id/members", workspaceHandler.ListMembers)
			workspaceGroup.PUT("/:id/members/:user_id", workspaceHandler.UpdateMemberRole)
			workspaceGroup.DELETE("/:id/members/:user_id", workspaceHandler.RemoveMember)
		}

		invitationGroup := v1.Group("/invitations")
		invitationGroup.Use(middleware.AuthMiddleware())
		{
			invitationGroup.GET("/me", workspaceHandler.ListMyInvitations)
			invitationGroup.POST("/:invite_id/accept", workspaceHandler.AcceptInvitation)
			invitationGroup.POST("/:invite_id/reject", workspaceHandler.RejectInvitation)
		}

		documentGroup := v1.Group("/documents")
		documentGroup.Use(middleware.AuthMiddleware())
		documentHandler := new(handler.DocumentHandler)
		collabHandler := new(handler.CollabHandler)
		aiProviderHandler := new(handler.AIProviderHandler)
		{
			documentGroup.POST("", documentHandler.Create)
			documentGroup.GET("", documentHandler.List)
			documentGroup.GET("/:id", documentHandler.GetByID)
			documentGroup.PATCH(":id", documentHandler.UpdateMeta)
			documentGroup.PUT("/:id", documentHandler.SaveVersion)
			documentGroup.DELETE(":id", documentHandler.Delete)
			documentGroup.GET("/:id/versions", documentHandler.VersionHistory)
			documentGroup.POST("/:id/restore", documentHandler.RestoreVersion)
			documentGroup.POST("/:id/collab-token", collabHandler.IssueToken)
			documentGroup.POST("/:id/collab/content", collabHandler.SyncContent)
		}

		aiProviderGroup := v1.Group("/ai/providers")
		aiProviderGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			aiProviderGroup.GET("", aiProviderHandler.List)
			aiProviderGroup.GET("/vendors", aiProviderHandler.ListVendors)
			aiProviderGroup.POST("/vendors", aiProviderHandler.CreateVendor)
			aiProviderGroup.PATCH("/vendors/:providerId/enabled", aiProviderHandler.UpdateVendorEnabled)
			aiProviderGroup.GET("/vendors/:providerId", aiProviderHandler.GetVendorDetail)
			aiProviderGroup.GET("/vendors/:providerId/discover-models", aiProviderHandler.DiscoverVendorModels)
			aiProviderGroup.POST("/vendors/:providerId/models", aiProviderHandler.CreateModelByVendor)
			aiProviderGroup.POST("", aiProviderHandler.Create)
			aiProviderGroup.GET("/:model", aiProviderHandler.GetByModel)
			aiProviderGroup.PUT("/:model", aiProviderHandler.Update)
			aiProviderGroup.DELETE("/:model", aiProviderHandler.Delete)
		}

		adminUserGroup := v1.Group("/admin")
		adminUserGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		adminUserHandler := new(handler.AdminUserHandler)
		{
			adminUserGroup.GET("/stats", adminUserHandler.GetStats)
			adminUserGroup.GET("/users", adminUserHandler.ListUsers)
			adminUserGroup.GET("/users/:userId", adminUserHandler.GetUser)
			adminUserGroup.PATCH("/users/:userId", adminUserHandler.UpdateUser)
			adminUserGroup.DELETE("/users/:userId", adminUserHandler.DeleteUser)
			adminUserGroup.POST("/users/:userId/reset-password", adminUserHandler.ResetPassword)
		}

		// 协作文档 WebSocket 使用短期 token 鉴权，不复用 AuthMiddleware。
		v1.GET("/documents/:id/collab/ws", collabHandler.ConnectWS)
	}
}
