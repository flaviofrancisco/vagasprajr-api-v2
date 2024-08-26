package jobs

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"math/rand"

	"github.com/google/uuid"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAggregatedJobsValues(body JobFilter) (JobFilterOptions, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return JobFilterOptions{}, err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return JobFilterOptions{}, err
	}	

	db := client.Database(mongodb_database)
	collection := db.Collection("jobs")

	companies, err := GetJobsAggregatedValues(collection, body, "company_name")
	locations, errLocation := GetJobsAggregatedValues(collection, body, "location")
	salaries, errSalaries := GetJobsAggregatedValues(collection, body, "salary")
	providers, errProviders := GetJobsAggregatedValues(collection, body, "provider")

	if err != nil || errLocation != nil || errSalaries != nil || errProviders != nil {
		return JobFilterOptions{}, err
	}

	return JobFilterOptions{
		Companies: companies,
		Locations: locations,
		Salaries:  salaries,
		Providers: providers,
	}, nil
}

func GetJobsAggregatedValues(collection *mongo.Collection, body JobFilter, field string) ([]string, error) {
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

	andConditions = append(andConditions, bson.M{field: bson.M{"$ne": nil}})

	filter["$and"] = andConditions


	results, err := collection.Distinct(context.Background(), field, filter)

	if err != nil {
		return nil, err
	}

	options := make([]string, 0, len(results))

	for _, result := range results {
		if companyName, ok := result.(string); ok {
			// check if companyName is null
			if companyName == "" {
				continue
			}
			options = append(options, companyName)
		} else {
			// Handle the case where the result is not a string
			return nil, fmt.Errorf("unexpected type for company name: %T", result)
		}
	}

	return options, nil	
}

func CreateJob(body CreateJobBody) (Job, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return Job{}, err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return Job{}, err
	}	

	db := client.Database(mongodb_database)
	collection := db.Collection("jobs")
	
	job := Job{
		Id:          uuid.New().String(),
		Title:       body.Title,
		Company:     body.Company,
		Location:    body.Location,
		Url:         body.Url,
		Salary:      body.Salary,
		Provider:    body.Provider,
		CreatedAt:   commons.GetBrasiliaTime(),
		JobDate:     commons.GetBrasiliaTime().Format(time.DateTime),
		Creator:     body.Creator,
		IsApproved:  false,
		IsClosed:    false,
		PostedOnDiscord:       false,
		PostedOnTelegram:      false,
		PostedOnMastodon:      false,
		PostedOnFacebook:      false,
		PostedOnTwitter:       false,
		PostedOnBlueSky:       false,		
	}

	shortUrl, detailUrl, code := CreateShortUrl()

	for {
		isAvailable, err := IsShortUrlAvailable(shortUrl)
		if err != nil {
			// handle error
			log.Fatalf("Error checking URL availability: %v", err)
		}
		if isAvailable {
			break
		}
		shortUrl, detailUrl, code = CreateShortUrl()
	}	

	job.Code = code
	job.JobShortUrl = shortUrl
	job.JobDetailsUrl = detailUrl

	if body.Provider == "" {
		job.Provider = "vagasprajr"
	}

	_, err = collection.InsertOne(context.Background(), job)

	if err != nil {
		return Job{}, err
	}

	return job, nil
}

func IsShortUrlAvailable(shortUrl string) (bool, error) {

	mongodb_database := os.Getenv("MONGODB_DATABASE")

	client, err := models.Connect()

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return false, err
	}

	collection := client.Database(mongodb_database).Collection("jobs")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result JobItem

	err = collection.FindOne(ctx, bson.M{"job_short_url": shortUrl}).Decode(&result)

	if err != nil {
		return true, nil
	}

	return false, nil
}

func GenerateCode() string {
	
	// Create a randon string with alfanumeric characters with 6 characters

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}

	code := string(b)

	return code
}

func CreateShortUrl() (string, string, string) {
 	
	code := GenerateCode()

	return os.Getenv("BASE_UI_HOST")+`/go/` + code, os.Getenv("BASE_UI_HOST")+`/v/` + code, code
}

func GetValidationToken() string {

	// Create a randon string with alfanumeric characters with 6 characters

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}

	return string(b)
}

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

	andConditions = appendCondition(andConditions, "company_name", body.Company)
	andConditions = appendCondition(andConditions, "location", body.Location)
	andConditions = appendCondition(andConditions, "salary", body.Salary)
	andConditions = appendCondition(andConditions, "provider", body.Provider)

	andConditions = appendInCondition(andConditions, "_id", body.Ids)
	andConditions = appendInCondition(andConditions, "company_name", body.JobFilterOptions.Companies)
	andConditions = appendInCondition(andConditions, "location", body.JobFilterOptions.Locations)
	andConditions = appendInCondition(andConditions, "provider", body.JobFilterOptions.Providers)
	andConditions = appendInCondition(andConditions, "salary", body.JobFilterOptions.Salaries)

	andConditions = append(andConditions, bson.M{"is_approved": true})
	andConditions = append(andConditions, bson.M{"is_closed": false})

	if body.CreatorId != primitive.NilObjectID {
		andConditions = append(andConditions, bson.M{"creator": body.CreatorId})
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

func GetOriginalURL(shortUrl string) (string, error) {

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

	collection := client.Database(mongodb_database).Collection("jobs")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result JobItem

	log.Print("[MONGODB]: Searching for: ", shortUrl)

	err = collection.FindOne(ctx, bson.M{"job_short_url": shortUrl, "is_approved": true, "is_closed": false }).Decode(&result)

	if err != nil {
		code := shortUrl[len(shortUrl)-6:]
		return os.Getenv("BASE_UI_HOST")+`/vagas/` + code, err
	}

	return result.Url, nil
}

func UpdateJobClicks(shortUrl string) error {

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

	collection := client.Database(mongodb_database).Collection("jobs")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.UpdateOne(ctx, bson.M{"job_short_url": shortUrl}, bson.M{"$inc": bson.M{"qty_clicks": 1}})

	if err != nil {
		return err
	}

	return nil
}

func appendCondition(andConditions []bson.M, field string, value string) []bson.M {
    if value != "" {
        andConditions = append(andConditions, bson.M{field: bson.M{"$regex": value, "$options": "i"}})
    }
    return andConditions
}

func appendInCondition(andConditions []bson.M, field string, values []string) []bson.M {
    if len(values) > 0 {
        andConditions = append(andConditions, bson.M{"$and": []bson.M{
            {field: bson.M{"$in": values}},
        }})
    }
    return andConditions
}