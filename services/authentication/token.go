package authentication

import (
	"context"
	"errors"
	"os"
	"time"

	"math/rand"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	USERS_TOKENS_COLLECTION = "users_tokens"
)

const (
	TOKEN_NAME = "vagasprajr_token"
)

type Token struct {
	Id             primitive.ObjectID `bson:"_id" json:"id"`
	UserId         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token          string             `bson:"token" json:"token"`
	ExpirationDate primitive.DateTime `bson:"expiration_date" json:"expiration_date"`
	CreatedAt      primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt      primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

func GetValidationToken() string {

	// Create a randon string with alfanumeric characters with 6 characters

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}

	return string(b)
}

func (tk *Token) GetToken(userInfo users.UserInfo, expirationTime time.Time) (string, error) {

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

	tk.ExpirationDate = primitive.NewDateTimeFromTime(expirationTime)

	token, err := t.SignedString(key)
	if err != nil {
		return "", errors.New(`{"error": "Error getting the token - ` + err.Error() + `"}`)
	}
	return token, nil
}

func ValidateStringToken(token string) (users.UserInfo, error) {
	
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

	userInfo := GetUserInfoFromClaims(claim)

	if time.Now().UTC().Unix() > int64(claim["exp"].(float64)) {
		return users.UserInfo{}, errors.New(`{"error": "Token expired"}`)
	}

	return userInfo, nil
}

func ValidateToken(context *gin.Context) (users.UserInfo, error) {

	token, _ := context.Cookie("token")

	if token == "" {

		token = context.GetHeader("Authorization")
		token = token[7:]
		
		if token == "" {
			return users.UserInfo{}, errors.New(`{"error": "Token not found"}`)
		}
	}

	if token == "" {
		return users.UserInfo{}, errors.New(`{"error": "Token not found"}`)
	}

	return ValidateStringToken(token)
}

func GetUserInfoFromClaims(jwtClaims jwt.MapClaims) users.UserInfo {
	var userInfo users.UserInfo

	userInfo.FirstName = jwtClaims["first_name"].(string)
	userInfo.LastName = jwtClaims["last_name"].(string)
	userInfo.Email = jwtClaims["email"].(string)
	userInfo.UserName = jwtClaims["user_name"].(string)
	userInfo.Id = jwtClaims["id"].(string)

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

	return userInfo
}

func (tk *Token) SetTokenCookie(c *gin.Context) {

	if tk.ExpirationDate.Time().IsZero() {
		panic("Expiration date is required")
	}

    secure := os.Getenv("COOKIE_SECURE") == "true"
    c.SetCookie(
        "vagasprajr_token",
        tk.Token,
        int(tk.ExpirationDate.Time().UTC().Unix()),
        "/",
        os.Getenv("COOKIE_DOMAIN"),
        secure,
        true,
    )
}

func (tk *Token) DeleteTokenCookie(c *gin.Context) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetCookie(
		"vagasprajr_token",
		"",
		-1,
		"/",
		os.Getenv("COOKIE_DOMAIN"),
		secure,
		true,
	)
}

func (tk *Token) SaveRefreshToken(userInfo users.UserInfo) error {

	expirationDate := time.Now().UTC().Add(time.Duration(24) * time.Hour)
	tk.ExpirationDate = primitive.NewDateTimeFromTime(expirationDate)
	currentDateTimeUTC := time.Now().UTC()
	
	token_string, err := tk.GetToken(userInfo, expirationDate)	

	if (err != nil) {
		return err
	}

	tk.Token = token_string

	user_id, err := primitive.ObjectIDFromHex(userInfo.Id)

	if err != nil {
		return err
	}

	tk.UserId = user_id

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
	filter := bson.D{{Key: "user_id", Value: tk.UserId}}

	user_token := users.UserToken{}
	
	err = db.Collection(USERS_TOKENS_COLLECTION).FindOne(context.Background(), filter).Decode(&user_token)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			
			user_token.Id = primitive.NewObjectID()
			user_token.CreatedAt = primitive.NewDateTimeFromTime(currentDateTimeUTC)			
			user_token.UpdatedAt = primitive.NewDateTimeFromTime(currentDateTimeUTC)
			user_token.ExpirationDate = tk.ExpirationDate
			user_token.UserId = tk.UserId
			user_token.Token = tk.Token

			_, err = db.Collection(USERS_TOKENS_COLLECTION).InsertOne(context.Background(), user_token)

			if err != nil {
				return err
			}

		} else {
			return err
		}
	} else {
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "token", Value: tk.Token},
			{Key: "expiration_date", Value: tk.ExpirationDate},
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