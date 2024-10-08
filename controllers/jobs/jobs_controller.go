package jobs

import (
	"net/http"

	"github.com/flaviofrancisco/vagasprajr-api-v2/middlewares"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/jobs"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"

	"github.com/gin-gonic/gin"
)

func DeleteJob(context *gin.Context) {
	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	code := context.Param("code")

	if code == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	err := jobs.DeleteJob(code)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func UpdateJob(context *gin.Context) {
	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	code := context.Param("code")	

	if code == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	var body jobs.UpdateJobBody
	context.BindJSON(&body)

	job, err := jobs.GetJob(code)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	job.IsApproved = body.IsApproved
	job.IsClosed = body.IsClosed

	result, err := jobs.UpdateJob(job)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, result)
}

func GetJobAsAdmin(context *gin.Context) {
	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.Writer.WriteHeaderNow()
		context.Status(http.StatusUnauthorized)  
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	code := context.Param("code")

	result, err := jobs.GetJob(code)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, result)
}

func GetJobsAsAdmin(context *gin.Context) {
	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.Writer.WriteHeaderNow()
		context.Status(http.StatusUnauthorized)  
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var body commons.FilterRequest
	context.BindJSON(&body)

	result, err := jobs.GetJobsAsAdmin(body)

	if err != nil {
		context.Writer.WriteHeaderNow()
		context.Status(http.StatusInternalServerError)  
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, result)

}

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


func GetJob(context *gin.Context) {
	code := context.Param("code")

	result, err := jobs.GetJob(code)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jobView := JobDetailView {
		Title: result.Title,
		Company: result.Company,
		Location: result.Location,
		Url: result.Url,
		Salary: result.Salary,
		Provider: result.Provider,
		Created_at: result.CreatedAt,
		Code: result.Code,
		Description: result.Description,		
	}
	
	context.JSON(http.StatusOK, jobView)
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