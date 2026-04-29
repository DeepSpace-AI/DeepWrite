package routes

import (
	"github.com/deepwrite/submission-service/internal/handler"
	"github.com/deepwrite/submission-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Setup(
	r *gin.Engine,
	journalHandler *handler.JournalHandler,
	submissionHandler *handler.SubmissionHandler,
	reviewHandler *handler.ReviewHandler,
	recommendHandler *handler.RecommendHandler,
	formatCheckHandler *handler.FormatCheckHandler,
	jwtAuth *middleware.JWTAuth,
) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	api := r.Group("/api")
	api.Use(jwtAuth.AuthMiddleware())
	{
		api.GET("/journals", journalHandler.List)
		api.GET("/journals/categories", journalHandler.GetCategories)
		api.GET("/journals/:id", journalHandler.Get)

		api.POST("/submissions/recommend", recommendHandler.RecommendJournals)
		api.POST("/submissions/check-format", formatCheckHandler.CheckFormat)

		api.POST("/submissions", submissionHandler.Create)
		api.GET("/submissions/project/:project_id", submissionHandler.ListByProject)
		api.GET("/submissions/:id", submissionHandler.Get)
		api.PUT("/submissions/:id", submissionHandler.Update)
		api.DELETE("/submissions/:id", submissionHandler.Delete)
		api.POST("/submissions/:id/submit", submissionHandler.SubmitToJournal)
		api.PUT("/submissions/:id/status", submissionHandler.UpdateStatus)
		api.POST("/submissions/:id/history", submissionHandler.AddHistory)

		api.GET("/submissions/:id/recommend", recommendHandler.RecommendForSubmission)
		api.POST("/submissions/:id/check-format", formatCheckHandler.CheckFormat)

		api.POST("/submissions/:id/reviews", reviewHandler.Create)
		api.GET("/submissions/:id/reviews", reviewHandler.ListBySubmission)
		api.PUT("/submissions/:id/reviews/:review_id", reviewHandler.Update)
		api.DELETE("/submissions/:id/reviews/:review_id", reviewHandler.Delete)
	}
}
