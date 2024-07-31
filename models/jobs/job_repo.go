package jobs

import (
	"context"
	"os"
	"strings"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetJobs(body JobFilter) (PaginatedResult, error) {

	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return PaginatedResult{}, err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return PaginatedResult{}, err
	}	

	db := client.Database(mongodb_database)
	collection := db.Collection("jobs")

	filter := bson.M{}
	andConditions := []bson.M{}

	if body.Title != "" {
		// Split the title by spaces
		words := strings.Fields(strings.TrimSpace(body.Title))

		// Create a slice of bson.M to store the conditions
		conditions := make([]bson.M, len(words))

		// Loop through the words and create a regex condition for each word
		for i, word := range words {
			word = commons.HandleValueForRegex(strings.TrimSpace(word))
			conditions[i] = bson.M{"title": bson.M{"$regex": word, "$options": "i"}}			
		}

		// Append the conditions to the andConditions slice
		andConditions = append(andConditions, bson.M{"$and": conditions})
	}	

	filter["$and"] = andConditions

	page := body.Page
	perPage := body.PageSize

	if (page - 1) < 0 {
		page = 1
	}

	skip := (page - 1) * perPage
	// Sort the documents by created_at in descending order
	options := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64(skip)).SetLimit(int64(perPage))
	cursor, err := collection.Find(context.Background(), filter, options)

	if err != nil {
		return PaginatedResult{}, err
	}

	var jobs []JobViewPublic

	if err = cursor.All(context.Background(), &jobs); err != nil {
		return PaginatedResult{}, err
	}

	total, err := collection.CountDocuments(context.Background(), filter)

	if err != nil {
		return PaginatedResult{}, err
	}

	defer cursor.Close(context.Background())

	return PaginatedResult{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Data:    jobs,
	}, nil

}