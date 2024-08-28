package shopping

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdReference struct {
	Id 				primitive.ObjectID `json:"id" bson:"_id"`
	Description 	string `json:"description" bson:"description"`
	IsActive 		bool `json:"is_active" bson:"is_active"`
	CreatedAt 		time.Time `json:"created_at" bson:"created_at"`
	ImageUrl 		string `json:"image_url" bson:"image_url"`
	Url 			string `json:"url" bson:"url"`
}

type AdReferencesPaginatedResult struct {
	Total   int64 
	Page    int
	PerPage int
	Data    []AdReference
}