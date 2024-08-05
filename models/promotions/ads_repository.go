package promotions

import (
	"context"
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"go.mongodb.org/mongo-driver/bson"
)

func UpdateAdvertisementClicks(shortUrl string) error {

	mongodb_database := os.Getenv("MONGODB_DATABASE")

	client, err := models.Connect()

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return err
	}

	collection := client.Database(mongodb_database).Collection("ads")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err != nil {
		return err
	}
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"short_url": shortUrl},
		bson.D{
			{"$inc", bson.D{{"qty_clicks", 1}}},
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func GetAdOriginalURL(shortUrl string) (string, error) {

	mongodb_database := os.Getenv("MONGODB_DATABASE")

	client, err := models.Connect()

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return "", err
	}

	collection := client.Database(mongodb_database).Collection("ads")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result AdItem

	err = collection.FindOne(ctx, bson.M{"short_url": shortUrl}).Decode(&result)

	if err != nil {
		return "", err
	}

	return result.OriginalUrl, nil
}