package shopping

import (
	"net/http"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/shopping"
	"github.com/gin-gonic/gin"
)

func GetAdReferences (context *gin.Context) {

	adReferences, err := shopping.GetAdReferences()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, adReferences)	
}

func GetFilteredAdReferences (context *gin.Context) {

	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.Writer.WriteHeaderNow()
		context.Status(http.StatusUnauthorized)  
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	var filter commons.FilterRequest
	context.BindJSON(&filter)

	adReferences, err := shopping.GetFilteredAdReferences(filter)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, adReferences)	
}

func GetAdReference (context *gin.Context) {
	
	id := context.Param("id")

	adReference, err := shopping.GetAdReference(id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, adReference)	
}

func UpdateAdReference (context *gin.Context) {
	
	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.Writer.WriteHeaderNow()
		context.Status(http.StatusUnauthorized)  
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	var adReference shopping.AdReference
	context.BindJSON(&adReference)

	err := shopping.UpdateAdReference(adReference)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Ad reference updated successfully"})	
}