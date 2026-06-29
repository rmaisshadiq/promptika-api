package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rmaisshadiq/critical-prompt-api/controllers"
	"github.com/rmaisshadiq/critical-prompt-api/middleware"
)

// SetupRoutes registers all API routes on the given Gin engine.
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")

	// --- Public routes (no auth required) ---
	auth := api.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// --- Protected routes (JWT required) ---
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	{
		// Session management
		sessions := protected.Group("/sessions")
		{
			sessions.POST("/start", controllers.StartSession)
			sessions.POST("/end", controllers.EndSession)
		}

		// Prompt classification
		protected.POST("/prompts", controllers.CreatePrompt)

		// Report generation
		reports := protected.Group("/reports")
		{
			reports.POST("/generate", controllers.GenerateReport)
			reports.GET("", controllers.GetReportHistory)
		}

		// Personal analytics dashboard
		users := protected.Group("/users")
		{
			users.GET("/dashboard-analytics", controllers.GetDashboardAnalytics)
		}
	}
}
