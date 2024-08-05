package roles

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/cache"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
)

func GetRoles() ([]Role, error) {

	roles, err := getRolesFromCache()

	if err != nil {
		return []Role{}, err
	}

	if len(roles) > 0 {
		return roles, nil
	}

	mongodb_database := os.Getenv("MONGODB_DATABASE")

	client, err := models.Connect()

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return []Role{}, err
	}

	collection := client.Database(mongodb_database).Collection("roles")	

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err != nil {
		return []Role{}, err
	}

	cursor, err := collection.Find(ctx, bson.M{})

	if err != nil {
		return []Role{}, err
	}
	
	if err = cursor.All(ctx, &roles); err != nil {
		return []Role{}, err
	}

	setCacheRoles(roles)

	return roles, nil
}

func setCacheRoles(roles []Role) error {
	redis, err := cache.NewRedis()

	if err != nil {
		return err
	}

	rolesBytes, err := json.Marshal(roles)

	if err != nil {
		return err
	}

	err = redis.RedisClient.Set("roles", rolesBytes, 0).Err()

	if err != nil {
		return err
	}

	return nil
}

func getRolesFromCache() ([]Role, error) {

	cacheServer, err := cache.NewRedis()	

	if err != nil {
		return []Role{}, err
	}

	// Get the roles from the cache
	rolesBytes, err := cacheServer.RedisClient.Get("roles").Bytes()

    if err != nil && err != redis.Nil {
        return []Role{}, err
    } else if err == redis.Nil {
		return []Role{}, nil
	}

	var roles []Role

	err = json.Unmarshal(rolesBytes, &roles)

	if err != nil {
		return []Role{}, err
	}

	return roles, nil

}