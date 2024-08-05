package roles

type Role struct {
	Id string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}