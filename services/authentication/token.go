package authentication

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	USERS_TOKENS_COLLECTION = "users_tokens"
)

type Token struct {
	Id             string             `bson:"_id"`
	UserId         primitive.ObjectID `bson:"user_id"`
	Token          string             `bson:"token"`
	ExpirationDate primitive.DateTime `bson:"expiration_date"`
	CreatedAt      primitive.DateTime `bson:"created_at"`
	UpdatedAt      primitive.DateTime `bson:"updated_at"`
}

func (tk *Token) GetRefreshToken(token string, expiration time.Time) (string, error) {
	var (
		key   []byte
		t     *jwt.Token
		claim jwt.MapClaims
	)

	key = []byte(os.Getenv("JWT_SECRET"))

	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return "", errors.New(`{"error": "Error parsing the token - ` + err.Error() + `"}`)
	}

	claim = t.Claims.(jwt.MapClaims)

	userInfo := GetUserInfoFromClaims(claim)

	if time.Now().Unix() > int64(claim["exp"].(float64)) {
		return "", errors.New(`{"error": "Token expired"}`)
	}

	var newToken string

	newToken, err = tk.GetToken(userInfo, false)

	if err != nil {
		return "", errors.New(`{"error": "Error getting the token - ` + err.Error() + `"}`)
	}

	return newToken, nil
}

func (tk *Token) GetToken(userInfo users.UserInfo, isRefreshToken bool) (string, error) {
	var (
		key   []byte
		t     *jwt.Token
		token string
	)

	var expiration time.Time

	if isRefreshToken {
		expiration = time.Now().Add(time.Minute * 60 * 24 * 7) // 7 days
	} else {
		expiration = time.Now().Add(time.Minute * 60) // 60 minutes
	}	

	key = []byte(os.Getenv("JWT_SECRET"))

	t = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         userInfo.Id,
		"first_name": userInfo.FirstName,
		"last_name":  userInfo.LastName,
		"user_name":  userInfo.UserName,
		"email":      userInfo.Email,
		"links":      userInfo.Links,
		"exp":        expiration.Unix(),
	})

	token, err := t.SignedString(key)
	if err != nil {
		return "", errors.New(`{"error": "Error getting the token - ` + err.Error() + `"}`)
	}
	return token, nil
}

func ValidateToken(context *gin.Context) (users.UserInfo, error) {

	token, err := context.Cookie("token")

	if token == "" {

		token = context.GetHeader("Authorization")
		token = token[7:]
		
		if token == "" {
			return users.UserInfo{}, errors.New(`{"error": "Token not found"}`)
		}
	}

	if token == "" || err != nil {
		return users.UserInfo{}, errors.New(`{"error": "Token not found"}`)
	}
	
	var (
		key   []byte
		t     *jwt.Token
		claim jwt.MapClaims
	)

	key = []byte(os.Getenv("JWT_SECRET"))

	t, err = jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return users.UserInfo{}, errors.New(`{"error": "Error parsing the token - ` + err.Error() + `"}`)
	}

	claim = t.Claims.(jwt.MapClaims)

	userInfo := GetUserInfoFromClaims(claim)

	if time.Now().Unix() > int64(claim["exp"].(float64)) {
		return users.UserInfo{}, errors.New(`{"error": "Token expired"}`)
	}

	return userInfo, nil
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
    secure := os.Getenv("COOKIE_SECURE") == "true"
    c.SetCookie(
        "token",
        tk.Token,
        60*24*30, // 30 days
        "/",
        os.Getenv("COOKIE_DOMAIN"),
        secure,
        true,
    )
}

func (tk *Token) DeleteTokenCookie(c *gin.Context) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetCookie(
		"token",
		"",
		-1,
		"/",
		os.Getenv("COOKIE_DOMAIN"),
		secure,
		true,
	)
}

func (tk *Token) SaveRefreshToken(userInfo users.UserInfo) error {
	
	token_string, err := tk.GetToken(userInfo, true)	

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
	
	err = db.Collection(USERS_TOKENS_COLLECTION).FindOne(context.Background(), filter).Decode(&tk)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			tk.Id = uuid.New().String()
			tk.CreatedAt = primitive.NewDateTimeFromTime(commons.GetBrasiliaTime())
			tk.UpdatedAt = primitive.NewDateTimeFromTime(commons.GetBrasiliaTime())

			_, err = db.Collection(USERS_TOKENS_COLLECTION).InsertOne(context.Background(), tk)

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
			{Key: "updated_at", Value: primitive.NewDateTimeFromTime(commons.GetBrasiliaTime())},
		}}}

		opts := options.Update().SetUpsert(true)
		_, err = db.Collection(USERS_TOKENS_COLLECTION).UpdateOne(context.Background(), filter, update, opts)

		if err != nil {
			return err
		}
	}
	return nil
}