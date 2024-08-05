package users

import (
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserProfileResponse struct {
	Id                   string                      `json:"id"`
	Email                string                      `json:"email"`
	UserName             string                      `json:"user_name"`
	AboutMe              string                      `json:"about_me"`
	FirstName            string                      `json:"first_name"`
	LastName             string                      `json:"last_name"`
	City                 string                      `json:"city"`
	State                string                      `json:"state"`
	Links                []users.UserLink           `json:"links"`
	IsEmailConfirmed     bool                        `json:"is_email_confirmed"`
	Roles                []primitive.ObjectID        `bson:"roles"`
	Experiences          []users.UserExperice       `json:"experiences"`
	IsPublic             bool                        `json:"is_public"`
	ProfileViews         int                         `json:"profile_views"`
	TechExperiences      []users.UserTechExperience `json:"tech_experiences"`
	Educations           []users.UserEducation      `json:"educations"`
	Certifications       []users.UserCertification  `json:"certifications"`
	JobPreference        users.UserJobPreference    `json:"job_preferences"`
	DiversityInfo        users.UserDiversityInfo    `json:"diversity_info"`
	IdiomsInfo           []users.UserIdiomInfo      `json:"idioms_info"`
	IsPublicForRecruiter bool                        `json:"is_public_for_recruiter"`
}