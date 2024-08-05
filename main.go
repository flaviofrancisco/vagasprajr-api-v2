package main

import (
	"log"
	"os"
	"regexp"

	"github.com/flaviofrancisco/vagasprajr-api-v2/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	// print the environment variables
	log.Println("DEBUG_MODE:", os.Getenv("DEBUG_MODE"))
	log.Println("VERSION:", os.Getenv("VERSION"))
	log.Println("BASE_UI_HOST:", os.Getenv("BASE_UI_HOST"))

	mongoDbUrl := os.Getenv("MONGODB_URL")

	if mongoDbUrl == "" {
		log.Fatal("MONGODB_URL is required")
	}

	re := regexp.MustCompile(`(mongodb://.*?:)(.*?)(@.*)`)
	modifiedURL := re.ReplaceAllString(mongoDbUrl, `$1*.*.*$3`)

	log.Println("MONGODB_URL:", modifiedURL)
	log.Println("MONGODB_DATABASE:", os.Getenv("MONGODB_DATABASE"))	

	server := gin.Default()

	routes.RegisterRoutes(server)
	
	server.Run(":3001")
}