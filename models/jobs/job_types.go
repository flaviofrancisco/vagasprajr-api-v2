package jobs

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaginatedResult struct {
	Total   int64 // Total number of filtered documents
	Page    int
	PerPage int
	Data    []JobViewPublic // Filtered documents
}

type JobViewPublic struct {
	Id          string    `json:"id" bson:"_id"`
	Title       string    `json:"title" bson:"title"`
	Company     string    `json:"company_name" bson:"company_name"`
	Location    string    `json:"location" bson:"location"`
	JobShortUrl string    `json:"job_short_url" bson:"job_short_url"`
	Salary      string    `json:"salary" bson:"salary"`
	QtyClicks   int       `json:"qty_clicks" bson:"qty_clicks"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	Provider    string     `json:"provider" bson:"provider"`
}

type JobFilter struct {
	Page      int                		`json:"page" bson:"page"`
	PageSize  int                		`json:"pageSize" bson:"pageSize"`
	Title     string             		`json:"title" bson:"title"`
	Company   string             		`json:"company_name" bson:"company_name"`
	Location  string             		`json:"location" bson:"location"`
	Salary    string             		`json:"salary" bson:"salary"`
	Provider  string             		`json:"provider" bson:"provider"`
	Ids       []string           		`json:"ids" bson:"ids"`
	CreatorId primitive.ObjectID 		`json:"creator_id" bson:"creator_id"`
	JobFilterOptions JobFilterOptions 	`json:"job_filter_options" bson:"job_filter_options"`
}

type JobFilterOptions struct {
	Companies []string           `json:"companies" bson:"companies"`
	Locations []string           `json:"locations" bson:"locations"`
	Providers []string           `json:"providers" bson:"providers"`
	Salaries  []string           `json:"salaries" bson:"salaries"`
}