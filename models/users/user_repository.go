package users

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/scrypt"
)

const (
	PW_SALT_BYTES = 32
	PW_HASH_BYTES = 64
)

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
	user.CreatedAt = primitive.NewDateTimeFromTime(commons.GetBrasiliaTime())
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
	update := bson.D{{"$set", bson.D{{"last_login", primitive.NewDateTimeFromTime(commons.GetBrasiliaTime())}}}}

	_, err = db.Collection("users").UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return nil	
}

func GetUserById(id string) (User, error) {
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

	object_id, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return User{}, err
	}

	filter := bson.D{{"_id", object_id}}
	var result User
	err = db.Collection("users").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, nil
		}
	}

	return result, nil	
}