package routes

import (
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/jobs"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/shopping"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/shorturls"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/users"
	"github.com/flaviofrancisco/vagasprajr-api-v2/middlewares/authentication"
	"github.com/flaviofrancisco/vagasprajr-api-v2/middlewares/authorization"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/oauth/google"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {

	debug := os.Getenv("DEBUG_MODE")	

	allowedOrigins := []string{"https://vagasprajr.com", "https://vagasprajr.com.br", "https://vagasparajr.com", "https://vagasparajr.com.br", "https://api.vagasprajr.com"}

	if debug == "true" {
		allowedOrigins = []string{"http://localhost:3000", "http://localhost:3001", "http://www.localhost:3000"}
	}

    server.Use(cors.New(cors.Config{
        AllowOrigins:     allowedOrigins,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))	


	// Health check
	server.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Admin routes
	admin := server.Group("/admin")
	admin.Use(authentication.AuthMiddleware(), authorization.AuthorizationMiddleware([]string{controllers.ADMIN}))
	admin.POST("/users", users.GetUsers)
	admin.GET("/users/:id", users.GetUser)

	// Authentication	
	server.POST("/auth/login", users.Login)	
	server.GET("/auth/logout", users.LogOut)		
	server.POST("/auth/forgotten-password", users.RequestPasswordReset)	
	server.POST("/auth/verify-reset-token", users.VerifyRessetToken)
	server.POST("/auth/reset-password", users.ResetPassword)	
	server.POST("/auth/authorize", authentication.AuthMiddleware(), users.IsAuthorized)	
	server.POST("/oauth/google", google.OAuthGoogle)
	
	// Jobs
	server.POST("/jobs/search", jobs.GetJobs)
	server.POST("/jobs/aggregated-values", jobs.GetAggregatedJobsValues)
	server.POST("/jobs", authentication.AuthMiddleware(), jobs.CreateJob)
	server.GET("/jobs/:code", jobs.GetJob)

	//Short URLs
	// Get the original job's URL from the short URL
	server.GET("/go/:code", shorturls.GetOriginalURL)	
	// Redirect to the ad's original URL from the short URL
	server.GET("/r/:code", shorturls.RedirectToOriginalAdURL)
	// Redirect to the job's original URL from the short URL
	server.GET("/j/:code", shorturls.RedirectToOriginalJobUrl)

	//Shopping
	server.GET("/shopping/ad-references", shopping.GetAdReferences)

	// Singn Up
	server.POST("/auth/signup", users.SignUp)
	server.GET("/auth/signup/confirm-email/:token", users.ConfirmEmail)
	
	// Token
	server.GET("/auth/refresh-token", users.RefreshToken)

	// Users	
	server.GET("/users/profile", authentication.AuthMiddleware(), users.GetUserProfile)
	server.PUT("/users/profile", authentication.AuthMiddleware(), users.UpdateUser)
	server.POST("/users/username", authentication.AuthMiddleware(), users.UpdateUserName)
}