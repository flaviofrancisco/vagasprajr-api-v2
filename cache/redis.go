package cache

import (
	"errors"
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
    
    _, err := client.Ping().Result()
    if err != nil {
        return Redis{}, errors.New("failed to connect to Redis")
    }

    return Redis{
        RedisClient: *client,
    }, nil
}