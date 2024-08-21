package cache

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

type Redis struct {
 RedisClient redis.Client
}

func NewRedis() (Redis, error) {

    var client = redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_SERVER"),
        Password: os.Getenv("REDIS_PASSWORD"), 
		DB:	   0,
    })

    var redis_password string
    
    if len(os.Getenv("REDIS_PASSWORD")) > 4 {
        redis_password = os.Getenv("REDIS_PASSWORD")[0:4]
    }

    fmt.Println("Connecting to Redis server: ", os.Getenv("REDIS_SERVER") + " with password: " + redis_password)

    _, err := client.Ping().Result()
    if err != nil {
        return Redis{}, errors.New("failed to connect to Redis")
    }

    return Redis{
        RedisClient: *client,
    }, nil
}