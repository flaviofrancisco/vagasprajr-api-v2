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

type JobItem struct {
	JobShortUrl string `bson:"job_short_url"`
	Url         string `bson:"url"`
}

type CreateJobBody struct {
	Id 			string			 	`json:"id" bson:"_id"`
	Title 		string 				`json:"title" bson:"title"`
	Company 	string 				`json:"company_name" bson:"company_name"`
	Location 	string 				`json:"location" bson:"location"`
	Url 		string 				`json:"url" bson:"url"`
	Salary 		string 				`json:"salary" bson:"salary"`
	Provider 	string 				`json:"provider" bson:"provider"`
	Created_at 	time.Time 			`json:"created_at" bson:"created_at"`	
	Code 		string 				`json:"code" bson:"code"`
	Creator		primitive.ObjectID  `json:"creator" bson:"creator"`
}

type Job struct {
	Id                    string                  `json:"id" bson:"_id"`
	Title                 string                  `json:"title" bson:"title"`
	Function              string                  `json:"function" bson:"function"`
	Industry              string                  `json:"industry" bson:"industry"`
	Url                   string                  `json:"url" bson:"url"`
	Description           string                  `json:"description" bson:"description"`
	Company               string                  `json:"company_name" bson:"company_name"`
	Remote                string                  `json:"home_office" bson:"home_office"`
	JobDate               string                  `json:"job_date" bson:"job_date"`
	CreatedAt             time.Time               `json:"created_at" bson:"created_at"`
	Location              string                  `json:"location" bson:"location"`
	Salary                string                  `json:"salary" bson:"salary"`
	PostedOnBlueSky       bool                    `json:"posted_on_bluesky" bson:"posted_on_bluesky"`
	PostedOnDiscord       bool                    `json:"posted_on_discord" bson:"posted_on_discord"`
	PostedOnTelegram      bool                    `json:"posted_on_telegram" bson:"posted_on_telegram"`
	PostedOnMastodon      bool                    `json:"posted_on_mastodon" bson:"posted_on_mastodon"`
	PostedOnFacebook      bool                    `json:"posted_on_facebook" bson:"posted_on_facebook"`
	PostedOnTwitter       bool                    `json:"posted_on_twitter" bson:"posted_on_twitter"`
	Provider              string                  `json:"provider" bson:"provider"`
	JobShortUrl           string                  `json:"job_short_url" bson:"job_short_url"`
	QtyClicks             int                     `json:"qty_clicks" bson:"qty_clicks"`
	IsApproved            bool                    `json:"is_approved" bson:"is_approved"`
	JobDetailsUrl         string                  `json:"job_details_url" bson:"job_details_url"`
	Creator               primitive.ObjectID      `json:"creator" bson:"creator"`
	IsClosed              bool                    `json:"is_closed" bson:"is_closed"`
	ClosedAt              time.Time               `json:"closed_at" bson:"closed_at"`
	Code                  string                  `json:"code" bson:"code"`
	AffirmativeParameters AffirmativeJobParameter `json:"affirmative_parameters" bson:"affirmative_parameters"`
	ContractType          string                  `json:"contract_type" bson:"contract_type"`
	LastUpdate            time.Time               `json:"last_update" bson:"last_update"`
	UdatedBy              primitive.ObjectID      `json:"updated_by" bson:"updated_by"`
}

type AffirmativeJobParameter struct {
	IsBlackPerson            bool `json:"is_black_person" bson:"is_black_person"`
	IsWomen                  bool `json:"is_women" bson:"is_women"`
	IsLgbtqia                bool `json:"is_lgbtqia" bson:"is_lgbtqia"`
	IsIndigenous             bool `json:"is_indigenous" bson:"is_indigenous"`
	IsPersonWithDisabilities bool `json:"is_person_with_disabilities" bson:"is_person_with_disabilities"`
}