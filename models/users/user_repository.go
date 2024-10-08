package users

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/roles"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/emails"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/scrypt"
)

const (
	PW_SALT_BYTES = 32
	PW_HASH_BYTES = 64
)

const (
	USERS_TOKENS_COLLECTION = "users_tokens"
)

func SignUp(user User) (error) {

	err:= commons.ValidatePassword(user.Password)

	if err != nil {
		return err
	}

	user.Email = strings.ToLower(user.Email)	
	user.ValidationToken = commons.GetValidationToken()

	err = CreateUser(user)

	if err != nil {
		return err
	}

	go emails.SendEmail("", []string{user.Email}, "Confirmação de email", emails.GetWelcomeEmail(user.ValidationToken))

	return nil

}

func CreateUser(user User) (error) {
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

	db := client.Database(mongodb_database)
	
	user.Id = primitive.NewObjectID()
	user.SetSaltedPassword()
	user.IsDeleted = false
	user.IsEmailConfirmed = false
	user.CreatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	user.LastLogin = primitive.DateTime(0)
	user.IsPublic = false
	user.ProfileViews = 0
	user.Links = []UserLink{}
	user.Experiences = []UserExperice{}
	user.TechExperiences = []UserTechExperience{}

	_, err = db.Collection("users").InsertOne(context.Background(), user)

	if err != nil {
		return err
	}

	return nil	
}

func (user *User) SetSaltedPassword() error {

	salt := make([]byte, PW_SALT_BYTES)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return err
	}

	hash, err := scrypt.Key([]byte(user.Password), salt, 1<<14, 8, 1, PW_HASH_BYTES)
	if err != nil {
		return err
	}

	// Encode the salt and hash as Base64
	saltStr := base64.StdEncoding.EncodeToString(salt)
	hashStr := base64.StdEncoding.EncodeToString(hash)

	// Store the salt securely or return it as needed
	user.Salt = saltStr
	user.Password = hashStr

	return nil
}

func (user *User) IsAuthenticated() (bool, error) {
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

	db := client.Database(mongodb_database)

	// Check if a user with the given email exists
	filter := bson.D{
		{Key: "email", Value: strings.ToLower(user.Email)},
		{Key: "is_deleted", Value: false},
	}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
	}

	if result.Email == "" {
		return false, errors.New("usuário não encontrado, verifique o e-mail ou senha")
	}

	// Decode the salt from the string
	salt, err := base64.StdEncoding.DecodeString(result.Salt)
	if err != nil {
		return false, err
	}

	// Hash the password with the salt
	hash, err := scrypt.Key([]byte(user.Password), salt, 1<<14, 8, 1, PW_HASH_BYTES)
	if err != nil {
		return false, err
	}

	// Encode the hash as Base64
	hashStr := base64.StdEncoding.EncodeToString(hash)

	if hashStr != result.Password {
		return false, errors.New("usuário não encontrado, verifique o e-mail ou senha")
	}

	return true, nil	
}

func GetUserByEmail(email string) (User, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return User{}, err
	}

	db := client.Database(mongodb_database)

	// Check if a user with the given email exists
	filter := bson.D{{Key: "email", Value: strings.ToLower(email)}}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, nil
		}
	}

	return result, nil	
}

func (user *User) UpdateLastLogin() error {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)

	filter := bson.D{{"email", strings.ToLower(user.Email)}}
	update := bson.D{{"$set", bson.D{{"last_login", primitive.NewDateTimeFromTime(time.Now().UTC())}}}}

	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil	
}

func Update(user *User) error {

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

	db := client.Database(mongodb_database)

	if user.JobPreference.JobLocations == nil || len(user.JobPreference.JobLocations) == 0 {
		user.JobPreference.JobLocations = []UserJobLocation{
			{
				City:     user.City,
				State:    user.State,
				Priority: 1,
			},
		}
	}

	filter := bson.D{{Key: "_id", Value: user.Id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "first_name", Value: user.FirstName},
			{Key: "last_name", Value: user.LastName},
			{Key: "city", Value: user.City},
			{Key: "state", Value: user.State},
			{Key: "links", Value: user.Links},
			{Key: "experiences", Value: user.Experiences},
			{Key: "user_name", Value: strings.ToLower(user.UserName)},
			{Key: "last_update", Value: user.LastUpdate},
			{Key: "about_me", Value: user.AboutMe},
			{Key: "is_public", Value: user.IsPublic},
			{Key: "tech_experiences", Value: user.TechExperiences},
			{Key: "educations", Value: user.Educations},
			{Key: "certifications", Value: user.Certifications},
			{Key: "job_preferences", Value: user.JobPreference},
			{Key: "diversity_info", Value: user.DiversityInfo},
			{Key: "idioms_info", Value: user.IdiomsInfo},
			{Key: "is_public_for_recruiter", Value: user.IsPublicForRecruiter},
		}},
	}

	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (user *User) IncrementProfileViews() error {
	
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

	filter := bson.D{{Key: "_id", Value: user.Id}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "public_profile_views", Value: 1}}}}

	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (user *User) ConfirmEmail() error {

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

	db := client.Database(mongodb_database)

	filter := bson.D{{Key: "_id", Value: user.Id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "is_email_confirmed", Value: true}}}}

	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func GetUserByUserName(userName string) (User, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return User{}, err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)

	filter := bson.D{
		{Key: "user_name", Value: strings.ToLower(userName)},
		{Key: "is_deleted", Value: false},
		{Key: "is_email_confirmed", Value: true},
		{Key: "is_public", Value: true},
	}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, nil
		}
	}

	return result, nil	
}

func (user *User) UpdateUserBookmarkedJobs() error {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)
	
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "_id", Value: user.Id}}

	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "bookmarked_jobs", Value: user.BookmarkedJobs}}},
		{Key: "$set", Value: bson.D{{Key: "last_update", Value: primitive.NewDateTimeFromTime(time.Now().UTC())}}},
	}
	
	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func DeleteUser(id primitive.ObjectID) error {
	
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)

	filter := bson.D{{Key: "_id", Value: id}}	

	_, err = db.Collection("users").DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	filter = bson.D{{Key: "user_id", Value: id}}
	_, err = db.Collection("users_tokens").DeleteMany(context.Background(), filter)

	if err != nil {
		return err
	}

	return nil
}

func GetUserById(id primitive.ObjectID) (User, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return User{}, err
	}

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)
	
	if err != nil {
		return User{}, err
	}

	filter := bson.D{{Key: "_id", Value: id}}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, nil
		}
	}

	return result, nil	
}

func GetUserRoles (id primitive.ObjectID) ([]string, error) {
	
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return []string{}, err
	}
	
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)

	filter := bson.D{{Key: "_id", Value: id}}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
	}

	if result.Roles == nil {
		return []string{}, nil
	}
	
	roles, err := roles.GetRoles()

	if err != nil {
		return []string{}, err
	}
	
	resultRoles := []string{}

	for _, role := range result.Roles {
		for _, r := range roles {
			roleID, err := primitive.ObjectIDFromHex(r.Id)
			if err != nil {
				continue // Skip invalid ObjectID
			}
			if roleID.Hex() == primitive.ObjectID(role).Hex() {
				resultRoles = append(resultRoles, r.Name)
			}
		}
	}
	
	return resultRoles, nil	
}

func GetUserByValidationToken(token string) (User, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return User{}, err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)

	filter := bson.D{{Key: "validation_token", Value: token}}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, nil
		}
	}

	return result, nil	
}

func GetUsers(filter commons.FilterRequest) (UsersPaginatedResult, error) {
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return UsersPaginatedResult{}, err
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

	cursor, err := db.Collection("users").Find(context.Background(), filter.GetFilter(), options)

	if err != nil {
		return UsersPaginatedResult{}, err
	}

	var users []UserView

	if err = cursor.All(context.Background(), &users); err != nil {
		return UsersPaginatedResult{}, err
	}

	total, err := db.Collection("users").CountDocuments(context.Background(), filter.GetFilter())

	return UsersPaginatedResult{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Data:    users,
	}, nil		
}

func UpdateValidationToken(user User) error {
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

	filter := bson.D{{Key: "_id", Value: user.Id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "validation_token", Value: user.ValidationToken}}}}

	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil	
}

func (user *User) UpdatePassword(password string) error {
	
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

	user.Password = password
	user.SetSaltedPassword()

	filter := bson.D{{Key: "_id", Value: user.Id}}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: user.Password}, {Key: "password_salt", Value: user.Salt}}}}

	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (user *User) ResetValidationToken() error {
	
	user.ValidationToken = commons.GetValidationToken()

	err := UpdateValidationToken(*user)

	if err != nil {
		return err
	}

	return nil
}	

func (user *User) IsAuthorized(roles string) bool {
	
	if user.Roles == nil {
		return false
	}

	for _, role := range user.Roles {
		if role.Hex() == roles {
			return true
		}
	}

	return false	
}

func (user *User) Update() error {
	
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

	filter := bson.D{{Key: "_id", Value: user.Id}}

	update := bson.D{
		{
			Key: "$set", 
			Value: bson.D{
				{Key: "first_name", Value: user.FirstName}, 
				{Key: "last_name", Value: user.LastName}, 
				{Key: "city", Value: user.City}, 
				{Key: "state", Value: user.State},
				{Key: "links", Value: user.Links},
				{Key: "last_update", Value: primitive.NewDateTimeFromTime(time.Now().UTC())},
				{Key: "about_me", Value: user.AboutMe},
				{Key: "tech_experiences", Value: user.TechExperiences},
				{Key: "idioms_info", Value: user.IdiomsInfo},
				{Key: "certifications", Value: user.Certifications},
				{Key: "educations", Value: user.Educations},
				{Key: "experiences", Value: user.Experiences},
				{Key: "is_public", Value: user.IsPublic},
				{Key: "profile_image_url", Value: user.ProfileImageUrl},
				{Key: "oauth_image_url", Value: user.OAuthImageURL},
			},
		},		
	}
	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func IsUserNameValid(userName string) bool {

	//Remove spaces from the user name
	userName = strings.Replace(userName, " ", "", -1)

	// The user name must have at least 3 alphanumeric characters. No special characters; symbols or spaces are allowed.
	if len(userName) < 3 {
		return false
	}
	
	//regex
	pattern := "^[a-zA-Z0-9]*$"
	match, _ := regexp.MatchString(pattern, userName)

	return match
}

func UserNameAlreadyExists(userName string) (bool, error) {
	
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return false, err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	db := client.Database(mongodb_database)

	filter := bson.D{{Key: "user_name", Value: strings.ToLower(userName)}}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
	}

	return true, nil
}

func (user *User) UpdateProfilePicture() error {
	
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

	filter := bson.D{{Key: "_id", Value: user.Id}}

	update := bson.D{
		{
			Key: "$set", 
			Value: bson.D{
				{Key: "profile_image_url", Value: user.ProfileImageUrl},
				{Key: "last_update", Value: primitive.NewDateTimeFromTime(time.Now().UTC())},
			},
		},		
	}
	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil		

}

func( user *User) UpdateUserName() error {
	
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

	filter := bson.D{{Key: "_id", Value: user.Id}}

	update := bson.D{
		{
			Key: "$set", 
			Value: bson.D{
				{Key: "user_name", Value: user.UserName},
				{Key: "last_update", Value: primitive.NewDateTimeFromTime(time.Now().UTC())},
			},
		},		
	}
	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func GetTalents(filter commons.FilterRequest, is_admin bool, is_recruiter bool) (UserTalentsPaginatedResult, error) {
	
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	if err != nil {
		return UserTalentsPaginatedResult{}, err
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

	final_filter := filter.GetFilter()
	
	is_public_filter := bson.M {
		"is_public": true,
	}

	if is_admin {
		is_public_filter = bson.M{}
	}

	if is_recruiter {
		is_public_filter = bson.M{
			"$or": []bson.M{
				{"is_public_for_recruiter": true},
				{"is_public": true},
			},
		}
	}

	if existingAnd, ok := final_filter["$and"]; ok {
		final_filter["$and"] = append(existingAnd.([]bson.M),is_public_filter)
	} else {
		final_filter["$and"] = []bson.M{
			is_public_filter,
		}
	}

	options := options.Find().SetSort(bson.M{filter.Sort: orderDirection}).SetSkip(int64(skip)).SetLimit(int64(perPage))

	cursor, err := db.Collection("users").Find(context.Background(), final_filter, options)

	if err != nil {
		return UserTalentsPaginatedResult{}, err
	}

	var users []UserTalentView

	if err = cursor.All(context.Background(), &users); err != nil {
		return UserTalentsPaginatedResult{}, err
	}

	total, err := db.Collection("users").CountDocuments(context.Background(), final_filter)

	return UserTalentsPaginatedResult{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Data:    users,
	}, nil	
}