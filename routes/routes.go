package routes

import (
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/jobs"
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

	server.POST("/jobs", jobs.GetJobs)
	server.POST("/jobs/aggregated-values", jobs.GetAggregatedJobsValues)
}