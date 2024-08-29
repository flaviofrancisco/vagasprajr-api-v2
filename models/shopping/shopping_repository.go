package shopping

import (
	"context"
	"os"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetFilteredAdReferences(filter commons.FilterRequest) (AdReferencesPaginatedResult, error) {
		collection:= "ad_references"

	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return AdReferencesPaginatedResult{}, err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)

	page:= filter.Page
	perPage:= filter.PageSize

	if (page - 1) < 0 {
		page = 1
	}

	skip := (page - 1) * perPage

	if filter.Sort == "" {
		filter.Sort = "created_at"
	}

	orderDirection := -1

	if filter.IsAscending {
		orderDirection = 1
	}

	options := options.Find().SetSort(bson.M{filter.Sort: orderDirection}).SetSkip(int64(skip)).SetLimit(int64(perPage))

	cursor, err := db.Collection(collection).Find(context.Background(), filter.GetFilter(), options)

	if err != nil {
		return AdReferencesPaginatedResult{}, err
	}

	var adReferences []AdReference

	if err = cursor.All(context.Background(), &adReferences); err != nil {
		return AdReferencesPaginatedResult{}, err
	}

	total, err := db.Collection(collection).CountDocuments(context.Background(), filter.GetFilter())

	return AdReferencesPaginatedResult{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Data:    adReferences,
	}, nil
}

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

func GetAdReference(id string) (AdReference, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return AdReference{}, err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)
	collection := db.Collection("ad_references")

	idHex, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return AdReference{}, err
	}

	filter := bson.M{
		"_id": idHex,
	}

	var adReference AdReference

	err = collection.FindOne(context.Background(), filter).Decode(&adReference)

	if err != nil {
		return AdReference{}, err
	}

	return adReference, nil
}

func CreateAdReference(adReference AdReference) error {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)
	collection := db.Collection("ad_references")

	adReference.Id = primitive.NewObjectID()
	adReference.CreatedAt =commons.GetBrasiliaTime()

	_, err = collection.InsertOne(context.Background(), adReference)

	if err != nil {
		return err
	}

	return nil
}

func DeleteAdReference(id string) error {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)
	collection := db.Collection("ad_references")

	idHex, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": idHex,
	}

	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return nil
}

func UpdateAdReference(adReference AdReference) error {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)
	collection := db.Collection("ad_references")

	filter := bson.M{
		"_id": adReference.Id,
	}

	update := bson.M{
		"$set": adReference,
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}