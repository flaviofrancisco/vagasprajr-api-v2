package routes

import (
	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/jobs"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	server.POST("/jobs", jobs.GetJobs)
}