package shopping

import "time"

type AdReference struct {
	Id string `json:"id" bson:"_id"`
	Description string `json:"description" bson:"description"`
	IsActive bool `json:"is_active" bson:"is_active"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	ImageUrl string `json:"image_url" bson:"image_url"`
}