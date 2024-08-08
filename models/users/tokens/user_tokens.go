package tokens

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserToken struct {
	Id             primitive.ObjectID  `bson:"_id"`
	UserId         primitive.ObjectID `bson:"user_id"`
	Token          string             `bson:"token"`
	ExpirationDate primitive.DateTime `bson:"expiration_date"`
	CreatedAt      primitive.DateTime `bson:"created_at"`
	UpdatedAt      primitive.DateTime `bson:"updated_at"`
}

func GetDateTimePlusSeconds (seconds int) time.Time {
	return time.Now().UTC().Add(time.Duration(seconds) * time.Second)
}

func GetDateTimePlus(minutes int) time.Time {
	return time.Now().UTC().Add(time.Duration(minutes) * time.Minute)
}

func GetDateTimePlusHours(hours int) time.Time {
	return time.Now().UTC().Add(time.Duration(hours) * time.Hour)
}

func getExpirationToken() time.Time {
	return GetDateTimePlusHours(1)
}

func getExpirationRefreshToken() time.Time {
	return GetDateTimePlusHours(24)
}

// SetToken - Set the token for the user with a default expiration time
func (userToken *UserToken) SetToken(userInfo users.UserInfo) (error)  {

	expirationDateTime := getExpirationToken()
	err := userToken.SetAuthenticationToken(userInfo, expirationDateTime)
	
	if err != nil {
		return err
	}

	return nil
}

// SetAuthenticationToken - Set the authentication token for the user with an expiration time
func (userToken *UserToken) SetAuthenticationToken(userInfo users.UserInfo, expirationDateTime time.Time) (error) {

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
		"exp":        expirationDateTime.Unix(),
	})

	userToken.ExpirationDate = primitive.NewDateTimeFromTime(expirationDateTime)

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

func ExtractUserInfoFromTokenString(token string) (users.UserInfo, error) {

	var (
		key   []byte		
		claim jwt.MapClaims
	)

	key = []byte(os.Getenv("JWT_SECRET"))

	t, _ := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})

	claim = t.Claims.(jwt.MapClaims)		
	
	return GetUserInfo(claim)
}


func GetUserInfoFromTokenString(token string) (users.UserInfo, error) {
	
	var (
		key   []byte		
		claim jwt.MapClaims
	)

	key = []byte(os.Getenv("JWT_SECRET"))

	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return users.UserInfo{}, errors.New(`{"error": "Error parsing the token - ` + err.Error() + `"}`)
	}

	claim = t.Claims.(jwt.MapClaims)
		
	if time.Now().UTC().Unix() > int64(claim["exp"].(float64)) {
		return users.UserInfo{}, errors.New(`{"error": "Token expired"}`)
	}

	return GetUserInfo(claim)
}

func ExtractUserInfoForTokenRefresh(context *gin.Context) (users.UserInfo, error) {
	
	token, _ := context.Cookie(TOKEN_NAME)

	if token == "" {

		token = context.GetHeader("Authorization")

		if token == "" {
			return users.UserInfo{}, nil
		}

		token = token[7:]
		
		if token == "" {
			return users.UserInfo{}, errors.New(`{"error": "Token not found"}`)
		}		
	}

	return ExtractUserInfoFromTokenString(token)
}

func GetUserInfoFromContext(context *gin.Context) (users.UserInfo, error) {

	token, err := context.Cookie(TOKEN_NAME)

	if err != nil {
		return users.UserInfo{}, errors.New(`{"error": "Token not found"}`)		
	}

	if token == "" {

		token = context.GetHeader("Authorization")

		if token == "" {
			return users.UserInfo{}, nil
		}

		token = token[7:]
		
		if token == "" {
			return users.UserInfo{}, errors.New(`{"error": "Token not found"}`)
		}
	}

	return GetUserInfoFromTokenString(token)
}

func GetUserInfo(jwtClaims jwt.MapClaims) (users.UserInfo, error) {

	userInfo := users.UserInfo{
		FirstName: jwtClaims["first_name"].(string),
		LastName:  jwtClaims["last_name"].(string),
		Email:     jwtClaims["email"].(string),
		UserName:  jwtClaims["user_name"].(string),
	}

    idStr, ok := jwtClaims["id"].(string)
    if !ok {
		return users.UserInfo{}, errors.New(`{"error": "Invalid ObjectID"}`)        
    }

    objectID, err := primitive.ObjectIDFromHex(idStr)

    if err != nil {
        return users.UserInfo{}, errors.New(`{"error": "Invalid ObjectID"}`)
    }

    userInfo.Id = objectID
	
	if jwtClaims["links"] != nil {
		links := jwtClaims["links"].([]interface{})
		userInfo.Links = make([]users.UserLink, len(links))
		for i, link := range links {
			linkMap := link.(map[string]interface{})
			userInfo.Links[i].Name = linkMap["name"].(string)
			userInfo.Links[i].Url = linkMap["url"].(string)
			userInfo.Links[i].Id = linkMap["id"].(float64)
		}
	} else {
		userInfo.Links = []users.UserLink{}
	}	

	return userInfo, nil
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

func SaveRefreshToken(userInfo users.UserInfo) error {

	userRefershToken := UserToken{
		UserId: userInfo.Id,
		ExpirationDate: primitive.NewDateTimeFromTime(getExpirationRefreshToken()),
	}

	userRefershToken.SetAuthenticationToken(userInfo, getExpirationRefreshToken())

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
	filter := bson.D{{Key: "user_id", Value: userInfo.Id}}

	fetched_user_token := UserToken{}
	
	err = db.Collection(users.USERS_TOKENS_COLLECTION).FindOne(context.Background(), filter).Decode(&fetched_user_token)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			
			fetched_user_token.Id = primitive.NewObjectID()
			fetched_user_token.CreatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())			
			fetched_user_token.UpdatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
			fetched_user_token.ExpirationDate = primitive.NewDateTimeFromTime(getExpirationRefreshToken())
			fetched_user_token.UserId = userRefershToken.UserId
			fetched_user_token.Token = userRefershToken.Token

			_, err = db.Collection(users.USERS_TOKENS_COLLECTION).InsertOne(context.Background(), fetched_user_token)

			if err != nil {
				return err
			}

		} else {
			return err
		}
	} else {
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "token", Value: userRefershToken.Token},
			{Key: "expiration_date", Value: userRefershToken.ExpirationDate},
			{Key: "updated_at", Value: primitive.NewDateTimeFromTime(time.Now().UTC())},
		}}}

		opts := options.Update().SetUpsert(true)
		_, err = db.Collection(users.USERS_TOKENS_COLLECTION).UpdateOne(context.Background(), filter, update, opts)

		if err != nil {
			return err
		}
	}
	return nil
}

func (u *UserToken) SetRefreshToken() error {

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

	filter := bson.D{{Key: "user_id", Value: u.UserId}}

	err = db.Collection(users.USERS_TOKENS_COLLECTION).FindOne(context.Background(), filter).Decode(&u)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			u.Token = ""
			return nil
		} else{
			return err
		}		
	}

	return nil
}