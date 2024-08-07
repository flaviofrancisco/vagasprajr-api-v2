package users

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (userToken *UserToken) SetToken(userInfo UserInfo, expirationTime time.Time) (error) {

	var (
		key   []byte
		t     *jwt.Token
		token string
	)

	key = []byte(os.Getenv("JWT_SECRET"))
	
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         userInfo.Id,
		"first_name": userInfo.FirstName,
		"last_name":  userInfo.LastName,
		"user_name":  userInfo.UserName,
		"email":      userInfo.Email,
		"links":      userInfo.Links,
		"exp":        expirationTime.Unix(),
	})

	userToken.ExpirationDate = primitive.NewDateTimeFromTime(expirationTime)

	token, err := t.SignedString(key)
	if err != nil {
		return errors.New(`{"error": "Error getting the token - ` + err.Error() + `"}`)
	}

	userToken.Token = token
	userToken.CreatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	userToken.UpdatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	return nil
}

const (
	TOKEN_NAME = "vagasprajr_token"
)

func GenerateValidationToken() string {
	
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}

	return string(b)
}

func (userInfo *UserInfo) SetUserInfoFromTokenString(token string) (error) {
	
	var (
		key   []byte		
		claim jwt.MapClaims
	)

	key = []byte(os.Getenv("JWT_SECRET"))

	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return errors.New(`{"error": "Error parsing the token - ` + err.Error() + `"}`)
	}

	claim = t.Claims.(jwt.MapClaims)
	
	userInfo.SetUserInfo(claim)

	if time.Now().UTC().Unix() > int64(claim["exp"].(float64)) {
		return errors.New(`{"error": "Token expired"}`)
	}

	return nil
}

func (userInfo *UserInfo) GetUserInfoFromContext(context *gin.Context) (error) {

	token, _ := context.Cookie("token")

	if token == "" {

		token = context.GetHeader("Authorization")
		token = token[7:]
		
		if token == "" {
			return errors.New(`{"error": "Token not found"}`)
		}
	}

	if token == "" {
		return errors.New(`{"error": "Token not found"}`)
	}

	return userInfo.SetUserInfoFromTokenString(token)
}

func (userInfo *UserInfo) SetUserInfo(jwtClaims jwt.MapClaims) error {
	userInfo.FirstName = jwtClaims["first_name"].(string)
	userInfo.LastName = jwtClaims["last_name"].(string)
	userInfo.Email = jwtClaims["email"].(string)
	userInfo.UserName = jwtClaims["user_name"].(string)

    idStr, ok := jwtClaims["id"].(string)
    if !ok {
		return errors.New(`{"error": "Invalid ObjectID"}`)        
    }

    objectID, err := primitive.ObjectIDFromHex(idStr)

    if err != nil {
        return errors.New(`{"error": "Invalid ObjectID"}`)
    }

    userInfo.Id = objectID
	
	if jwtClaims["links"] != nil {
		links := jwtClaims["links"].([]interface{})
		userInfo.Links = make([]UserLink, len(links))
		for i, link := range links {
			linkMap := link.(map[string]interface{})
			userInfo.Links[i].Name = linkMap["name"].(string)
			userInfo.Links[i].Url = linkMap["url"].(string)
			userInfo.Links[i].Id = linkMap["id"].(float64)
		}
	} else {
		userInfo.Links = []UserLink{}
	}	

	return nil
}

func (userToken *UserToken) SetTokenCookie(c *gin.Context) {

	if userToken.ExpirationDate.Time().IsZero() {
		panic("Expiration date is required")
	}

    secure := os.Getenv("COOKIE_SECURE") == "true"
    c.SetCookie(
        TOKEN_NAME,
        userToken.Token,
        int(userToken.ExpirationDate.Time().UTC().Unix()),
        "/",
        os.Getenv("COOKIE_DOMAIN"),
        secure,
        true,
    )
}

func DeleteTokenCookie(c *gin.Context) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetCookie(
		TOKEN_NAME,
		"",
		-1,
		"/",
		os.Getenv("COOKIE_DOMAIN"),
		secure,
		true,
	)
}

func SaveRefreshToken(userInfo UserInfo) error {

	userToken := UserToken{
		UserId: userInfo.Id,
	}
	
	expirationDate := time.Now().UTC().Add(time.Duration(24) * time.Hour)
	userToken.ExpirationDate = primitive.NewDateTimeFromTime(expirationDate)
	currentDateTimeUTC := time.Now().UTC()
	
	err := userToken.SetToken(userInfo, expirationDate)	

	if (err != nil) {
		return err
	}
	
	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return err
	}

	db := client.Database(mongodb_database)

	// Check if user already exists
	filter := bson.D{{Key: "user_id", Value: userToken.UserId}}

	fetched_user_token := UserToken{}
	
	err = db.Collection(USERS_TOKENS_COLLECTION).FindOne(context.Background(), filter).Decode(&fetched_user_token)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			
			fetched_user_token.Id = primitive.NewObjectID()
			fetched_user_token.CreatedAt = primitive.NewDateTimeFromTime(currentDateTimeUTC)			
			fetched_user_token.UpdatedAt = primitive.NewDateTimeFromTime(currentDateTimeUTC)
			fetched_user_token.ExpirationDate = userToken.ExpirationDate
			fetched_user_token.UserId = userToken.UserId
			fetched_user_token.Token = userToken.Token

			_, err = db.Collection(USERS_TOKENS_COLLECTION).InsertOne(context.Background(), fetched_user_token)

			if err != nil {
				return err
			}

		} else {
			return err
		}
	} else {
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "token", Value: userToken.Token},
			{Key: "expiration_date", Value: userToken.ExpirationDate},
			{Key: "updated_at", Value: primitive.NewDateTimeFromTime(currentDateTimeUTC)},
		}}}

		opts := options.Update().SetUpsert(true)
		_, err = db.Collection(USERS_TOKENS_COLLECTION).UpdateOne(context.Background(), filter, update, opts)

		if err != nil {
			return err
		}
	}
	return nil
}