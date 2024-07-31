package main

import (
	"fmt"
	"log"
	"os"

	"github.com/flaviofrancisco/vagasprajr-api-v2/controllers/jobs"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// print the environment variables
	fmt.Println("DEBUG_MODE:", os.Getenv("DEBUG_MODE"))
	fmt.Println("VERSION:", os.Getenv("VERSION"))
	fmt.Println("BASE_UI_HOST:", os.Getenv("BASE_UI_HOST"))

	server := gin.Default()

	server.POST("/jobs", jobs.GetJobs)	
	server.Run(":8080")
}