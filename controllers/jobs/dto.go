package jobs

import (
	"time"
)

type JobDetailView struct {	
	Title 		string 				`json:"title" bson:"title"`
	Company 	string 				`json:"company_name" bson:"company_name"`
	Location 	string 				`json:"location" bson:"location"`
	Url 		string 				`json:"url" bson:"url"`
	Salary 		string 				`json:"salary" bson:"salary"`
	Provider 	string 				`json:"provider" bson:"provider"`
	Created_at 	time.Time 			`json:"created_at" bson:"created_at"`	
	Code 		string 				`json:"code" bson:"code"`	
	Description string              `json:"description" bson:"description"`
}