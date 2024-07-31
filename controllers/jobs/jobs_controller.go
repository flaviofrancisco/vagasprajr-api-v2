package jobs

import (
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/jobs"
	"github.com/gin-gonic/gin"
)

func GetJobs(context *gin.Context) {
	var body jobs.JobFilter
	context.BindJSON(&body)

	result, err := jobs.GetJobs(body)

	if err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	context.JSON(200, result)
}