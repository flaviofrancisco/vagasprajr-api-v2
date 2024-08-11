package users

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"	
	"strings"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/roles"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/emails"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	filter := bson.D{{Key: "email", Value: strings.ToLower(user.Email)}}
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