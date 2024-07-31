package jobs

import (
	"github.com/gin-gonic/gin"
)

func getJobs(context *gin.Context) {
	context.JSON(200, gin.H{
		"message": "getJobs",
	})
}