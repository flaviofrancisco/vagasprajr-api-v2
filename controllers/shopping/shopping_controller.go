package shopping

import (
	"net/http"

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