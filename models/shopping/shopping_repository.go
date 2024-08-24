package shopping

import (
	"context"
	"os"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"go.mongodb.org/mongo-driver/bson"
)

func GetAdReferences() ([]AdReference, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return nil, err		
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()	

	db := client.Database(mongodb_database)
	collection := db.Collection("ad_references")

	filter := bson.M{
		"is_active": true,
	}

	cursor, err := collection.Find(context.Background(), filter)
	
	if err != nil {
		return nil, err
	}

	var adReferences []AdReference

	for cursor.Next(context.Background()) {
		var adReference AdReference
		err := cursor.Decode(&adReference)

		if err != nil {
			return nil, err
		}

		adReferences = append(adReferences, adReference)
	}

	return adReferences, nil
}