package jobs

import (
	"net/http"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/jobs"
	"github.com/gin-gonic/gin"
)

func GetJobs(context *gin.Context) {
	var body jobs.JobFilter
	context.BindJSON(&body)

	result, err := jobs.GetJobs(body)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, result)
}

func GetAggregatedJobsValues(context *gin.Context) {
	var body jobs.JobFilter
	context.BindJSON(&body)

	result, err := jobs.GetAggregatedJobsValues(body)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, result)
}