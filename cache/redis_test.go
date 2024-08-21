package cache

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
    // Load .env file
    err := godotenv.Load("../.env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Run tests
    code := m.Run()
    os.Exit(code)
}

func TestNewRedis(t *testing.T) {
    // Call the NewRedis function
    redisInstance, err := NewRedis()

    // Assert no error occurred
    assert.NoError(t, err)

    // Assert the Redis client is not nil
    assert.NotNil(t, redisInstance.RedisClient)    
}