package shorturls

import (
	"log"
	"net/http"
	"os"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/jobs"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/promotions"
	"github.com/gin-gonic/gin"
)

func GetOriginalURL(context *gin.Context) {

	code := context.Param("code")

	if (code == "") {
		context.JSON(400, gin.H{"error": "Code not found"})
		return
	}

	shortUrl := os.Getenv("BASE_UI_HOST") + `/go/` + code
	originalUrl, _ := jobs.GetOriginalURL(shortUrl)

	// Create a map to hold the original URL
	response := make(map[string]string)
	response["originalUrl"] = originalUrl

	log.Print("Fetched original URL: ", originalUrl)
	
	context.JSON(200, response)	
}

func RedirectToOriginalAdURL(context *gin.Context) {
	code := context.Param("code")

	if (code == "") {
		context.JSON(400, gin.H{"error": "Code not found"})
		return
	}

	shortUrl := os.Getenv("BASE_UI_HOST") + `/r/` + code	
	originalUrl, err := promotions.GetAdOriginalURL(shortUrl)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Problem getting the original URL"})
		return
	}

	err = promotions.UpdateAdvertisementClicks(shortUrl)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Problem updating the advertisement clicks"})
		return
	}

	log.Print("redirecting to: ", originalUrl)
		
	context.Redirect(http.StatusTemporaryRedirect, originalUrl)
}

func RedirectToOriginalJobUrl(context *gin.Context) {
	code := context.Param("code")

	if (code == "") {
		context.JSON(400, gin.H{"error": "Code not found"})
		return
	}

	shortUrl := os.Getenv("BASE_UI_HOST") + `/go/` + code	
	log.Print("looking for shortUrl: ", shortUrl)

	originalUrl, err := jobs.GetOriginalURL(shortUrl)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Problem getting the original URL"})
		return
	}

	err = promotions.UpdateAdvertisementClicks(shortUrl)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Problem updating the advertisement clicks"})
		return
	}

	log.Print("redirecting to: ", originalUrl)

	context.Redirect(http.StatusTemporaryRedirect, originalUrl)

}