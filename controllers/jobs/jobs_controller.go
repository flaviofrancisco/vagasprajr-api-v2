package jobs

import (
	"net/http"

	"github.com/flaviofrancisco/vagasprajr-api-v2/middlewares"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/jobs"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
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

func CreateJob(context *gin.Context) {

	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)	

	if !context_error {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if (currentUser == nil) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)

	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)

	if err != nil {	
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var body jobs.CreateJobBody
	context.BindJSON(&body)

	body.Creator = user.Id

	result, err := jobs.CreateJob(body)

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