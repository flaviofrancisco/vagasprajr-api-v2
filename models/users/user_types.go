package users

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken 			string   			`json:"access_token"`
	Success     			bool     			`json:"success"`
	UserInfo    			UserInfo 			`json:"user_info"`
	ExpirationDate 			primitive.DateTime 	`json:"expiration_date"`		
}

type UserInfo struct {
	Id              	primitive.ObjectID 		`bson:"_id" json:"id"`
	FirstName       	string            		`bson:"first_name" json:"first_name"`
	LastName        	string            		`bson:"last_name" json:"last_name"`
	Email           	string            		`bson:"email" json:"email"`
	Links           	[]UserLink 		  		`bson:"links" json:"links"`
	UserName        	string            		`bson:"user_name" json:"user_name"`
	ProfileImageUrl 	string            		`bson:"profile_image_url" json:"profile_image_url"`
	Provider        	string            		`bson:"provider" json:"provider"`
	City            	string            		`bson:"city" json:"city"`
	State           	string            		`bson:"state" json:"state"`
	CreatedAt           primitive.DateTime   	`bson:"created_at" json:"created_at"`
	LastUpdate          primitive.DateTime   	`bson:"last_update" json:"last_update"`
	LastLogin           primitive.DateTime   	`bson:"last_login" json:"last_login"`
	Experices			[]UserExperice 			`bson:"experiences" json:"experiences"`
	Educations			[]UserEducation 		`bson:"educations" json:"educations"`
	IdiomsInfo			[]UserIdiomInfo 		`bson:"idioms_info" json:"idioms_info"`
	UserTechExperiences []UserTechExperience 	`bson:"tech_experiences" json:"tech_experiences"`			
	UserCertifications  []UserCertification 	`bson:"certifications" json:"certifications"`
}

type UserView struct {
	Id              primitive.ObjectID 		  `bson:"_id" json:"id"`
	FirstName            string               `bson:"first_name" json:"first_name"`
	LastName             string               `bson:"last_name" json:"last_name"`
	Email                string               `bson:"email" json:"email"`
	UserName             string               `bson:"user_name" json:"user_name"`
	City                 string               `bson:"city" json:"city"`
	State                string               `bson:"state" json:"state"`
	CreatedAt            primitive.DateTime   `bson:"created_at" json:"created_at"`	
	LastLogin            primitive.DateTime   `bson:"last_login" json:"last_login"`
	LastUpdate           primitive.DateTime   `bson:"last_update" json:"last_update"`
}

type User struct {
	Id                   primitive.ObjectID   `bson:"_id" json:"id"`
	FirstName            string               `bson:"first_name" json:"first_name"`
	LastName             string               `bson:"last_name" json:"last_name"`
	City                 string               `bson:"city" json:"city"`
	State                string               `bson:"state" json:"state"`
	Email                string               `bson:"email" json:"email"`
	Password             string               `bson:"password" json:"password"`
	Salt                 string               `bson:"password_salt" json:"password_salt"`
	CreatedAt            primitive.DateTime   `bson:"created_at" json:"created_at"`
	IsDeleted            bool                 `bson:"is_deleted" json:"is_deleted"`
	IsBlocked            bool                 `bson:"is_blocked" json:"is_blocked"`
	BlockedAt            primitive.DateTime   `bson:"blocked_at" json:"blocked_at"`
	ReasonOfBlock        string               `bson:"reason_of_block" json:"reason_of_block"`
	IsEmailConfirmed     bool                 `bson:"is_email_confirmed" json:"is_email_confirmed"`
	Roles                []primitive.ObjectID `bson:"roles" json:"roles"`
	Provider             string               `bson:"provider" json:"provider"`
	LastLogin            primitive.DateTime   `bson:"last_login" json:"last_login"`
	ValidationToken      string               `bson:"validation_token" json:"validation_token"`
	BookmarkedJobs       []string             `bson:"bookmarked_jobs" json:"bookmarked_jobs"`
	Links                []UserLink           `bson:"links" json:"links"`
	Experiences          []UserExperice       `bson:"experiences" json:"experiences"`
	UserName             string               `bson:"user_name" json:"user_name"`
	LastUpdate           primitive.DateTime   `bson:"last_update" json:"last_update"`
	AboutMe              string               `bson:"about_me" json:"about_me"`
	IsPublic             bool                 `bson:"is_public" json:"is_public"`
	IsPublicForRecruiter bool                 `bson:"is_public_for_recruiter" json:"is_public_for_recruiter"`
	ProfileViews         int                  `bson:"public_profile_views" json:"public_profile_views"`
	TechExperiences      []UserTechExperience `bson:"tech_experiences" json:"tech_experiences"`
	Educations           []UserEducation      `bson:"educations" json:"educations"`
	Certifications       []UserCertification  `bson:"certifications" json:"certifications"`
	JobPreference        UserJobPreference    `bson:"job_preferences" json:"job_preferences"`
	DiversityInfo        UserDiversityInfo    `bson:"diversity_info" json:"diversity_info"`
	IdiomsInfo           []UserIdiomInfo      `bson:"idioms_info" json:"idioms_info"`
}

type UserTechExperience struct {
	Id               int    `json:"id"`
	Technology       string `json:"technology"`
	ExperienceNumber int    `json:"experience_number"`
	ExperienceTime   string `json:"experience_time"`
}

type UserEducation struct {
	Id           float64 `json:"id"`
	Institution  string  `json:"institution"`
	Course       string  `json:"course"`
	FieldOfStudy string  `json:"field_of_study"`
	Grade        string  `json:"grade"`
	Description  string  `json:"description"`
	Degree       string  `json:"degree"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
}

type UserCertification struct {
	Id             float64 `json:"id"`
	IssuingCompany string  `json:"issuing_company"`
	Name           string  `json:"name"`
	IssueDate      string  `json:"start_date"`
	ExpirationDate string  `json:"end_date"`
	CredentialId   string  `json:"credential_id"`
	CredentialUrl  string  `json:"credential_url"`
}

type UserProfileFilter struct {
	QuickSearch           string   `json:"quick_search"`
	Page                  int      `json:"page"`
	PageSize              int      `json:"page_size"`
	City                  string   `json:"city"`
	State                 string   `json:"state"`
	ContractTypes         []string `json:"contract_types"`
	JobModes              []string `json:"job_modes"`
	IsAvailableForTravels bool     `json:"is_available_for_travels"`
}

type UserDiversityInfo struct {
	HealthInfo UserHealthInfo `json:"health_info"`
	GenderInfo UserGenderInfo `json:"gender_info"`
	RaceInfo   UserRaceInfo   `json:"race_info"`
}

type UserIdiomInfo struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

type UserHealthInfo struct {
	HasHealthCondition  bool   `json:"has_health_condition"`
	HealthConditionName string `json:"health_condition_name"`
	HasMedicalReport    bool   `json:"has_medical_report"`
}

type UserGenderInfo struct {
	IsLgbtq bool   `json:"is_lgbtqia"`
	Pronoun string `json:"pronoun"`
}

type UserRaceInfo struct {
	IsIndigenous bool `json:"is_indigenous"`
	IsBlack      bool `json:"is_black"`
}

type UserLink struct {
	Id   float64 `json:"id"`
	Url  string  `json:"url"`
	Name string  `json:"name"`
}

type UserExperice struct {
	Id          float64 `json:"id"`
	Company     string  `json:"company"`
	Position    string  `json:"position"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	Description string  `json:"description"`
}

type UserJobPreference struct {
	JobLocations          []UserJobLocation `json:"job_locations"`
	JobContractTypes      []string      `json:"job_contract_types"`
	IsAvalaibleForTravels bool          `json:"is_avalaible_for_travels"`
	JobModes              []string      `json:"job_modes"`
	MinMonthlySalary      float64       `json:"min_monthly_salary"`
}

type UserJobLocation struct {
	City     string `json:"city"`
	State    string `json:"state"`
	Priority int    `json:"priority"`
}

type UsersRequest struct {
    Sort     	string  	`bson:"sort" json:"sort"`
	IsAscending bool 		`bson:"is_ascending" json:"is_ascending"`
    Page     	int      	`bson:"page" json:"page"`
    PageSize 	int      	`bson:"page_size" json:"page_size"`
    Filters  	[]Filter 	`bson:"filters" json:"filters"`
}

type Filter struct {
	Operator string `bson:"operator" json:"operator"`
	Fields []Field `bson:"fields" json:"fields"`
}

type Field struct {
	Name 		string `bson:"name" json:"name"`
	Value 		string `bson:"value" json:"value"`
	Type 		string `bson:"type" json:"type"`
	MinValue 	string `bson:"min_value" json:"min_value"`
	MaxValue 	string `bson:"max_value" json:"max_value"`
}

type UsersPaginatedResult struct {
	Total   int64 
	Page    int
	PerPage int
	Data    []UserView
}