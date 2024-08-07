package routes

import (
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/services/oauth/google"

	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/jobs"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/shorturls"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/tokens"
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/users"
	"github.com/flaviofrancisco/vagasprajr-api-v2/middlewares/authentication"
	"github.com/flaviofrancisco/vagasprajr-api-v2/middlewares/authorization"
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
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
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
	admin.GET("/users", users.GetUsers)

	// Jobs
	server.POST("/jobs/search", jobs.GetJobs)
	server.POST("/jobs/aggregated-values", jobs.GetAggregatedJobsValues)

	//Short URLs
	// Get the original job's URL from the short URL
	server.GET("/go/:code", shorturls.GetOriginalURL)	
	// Redirect to the ad's original URL from the short URL
	server.GET("/r/:code", shorturls.RedirectToOriginalAdURL)
	// Redirect to the job's original URL from the short URL
	server.GET("/j/:code", shorturls.RedirectToOriginalJobUrl)

	// Users	
	server.POST("/auth/login", users.Login)	
	server.GET("/auth/logout", users.LogOut)	
	server.POST("/oauth/google", google.OAuthGoogle)

	// Singn Up
	server.POST("/auth/signup", users.SignUp)
	server.GET("/auth/signup/confirm-email/:token", users.ConfirmEmail)
	
	// Token
	server.GET("/auth/refresh-token", tokens.GetRefreshToken)

	server.GET("/users/:id", authentication.AuthMiddleware(), users.GetUser)
}