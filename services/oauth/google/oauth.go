package google

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/authentication"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/emails"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

const (
	USERS_GOOGLE_COLLECTION = "users_google"
)

type GoogleTokenRequest struct {
	AccessToken string `json:"access_token"`
}

type GoogleUserInfo struct {
	ID         string             `json:"id"`
	Email      string             `json:"email"`
	Verified   bool               `json:"verified_email"`
	Name       string             `json:"name"`
	Picture    string             `json:"picture"`
	Locale     string             `json:"locale"`
	GivenName  string             `json:"given_name"`
	FamilyName string             `json:"family_name"`
	IsDeleted  bool               `json:"is_deleted"`
	CreatedAt  primitive.DateTime `json:"created_at"`
	UpdatedAt  primitive.DateTime `json:"updated_at"`
	DeletedAt  primitive.DateTime `json:"deleted_at"`
}

func OAuthGoogle(context *gin.Context) {
	
	var token_request GoogleTokenRequest
	err := context.BindJSON(&token_request)
	
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})		
		return
	}

	if token_request.AccessToken == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Token de acesso não informado"})		
		return
	}

	response, err := http.Get(oauthGoogleUrlAPI + token_request.AccessToken)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the response body
	var googleUserInfo GoogleUserInfo

	err = json.Unmarshal(contents, &googleUserInfo)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser, _ := users.GetUserByEmail(googleUserInfo.Email)

	if currentUser.Email == "" {

		new_google_user, err := googleUserInfo.Create()

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		new_user := users.User{
			Email:     new_google_user.Email,
			FirstName: new_google_user.GivenName,
			LastName:  new_google_user.FamilyName,
		}

		err = users.CreateUser(new_user)

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		currentUser, err = users.GetUserByEmail(new_google_user.Email)

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		currentUser.IsEmailConfirmed = googleUserInfo.Verified

		if !currentUser.IsEmailConfirmed {
			currentUser.ValidationToken = authentication.GetValidationToken()
			go emails.SendEmail("", []string{new_google_user.Email}, "Confirmação de email", emails.GetWelcomeEmail(currentUser.ValidationToken))
		}

		err = users.Update(&currentUser)

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

	} else {

		user_google, err := users.GetUserByEmail(googleUserInfo.Email)

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if user_google.Email == "" {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Usuário não encontrado"})
			return
		}



		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if googleUserInfo.Verified && !currentUser.IsEmailConfirmed {			
			err := currentUser.ConfirmEmail()
			if err != nil {
				context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	}

	currentUser.UpdateLastLogin()

	var userInfo users.UserInfo

	userInfo.Email = strings.ToLower(currentUser.Email)
	userInfo.FirstName = currentUser.FirstName
	userInfo.LastName = currentUser.LastName
	userInfo.Links = currentUser.Links
	userInfo.Provider = currentUser.Provider
	userInfo.UserName = currentUser.UserName
	userInfo.Id = currentUser.Id.Hex()
	
	expirationDate := time.Now().UTC().Add(time.Duration(1) * time.Hour)

	token := authentication.Token{}
	
	accessToken, err := token.GetToken(userInfo, expirationDate)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result:= users.AuthResponse{
		AccessToken: accessToken,
		Success: true,
		UserInfo: userInfo,
		ExpirationDate: token.ExpirationDate,		
	}

	token.SetTokenCookie(context)

	err = token.SaveRefreshToken(userInfo)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, result)
}

func (user *GoogleUserInfo) Create() (*GoogleUserInfo, error) {

	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return nil, err
	}

	db := client.Database(mongodb_database)

	user.CreatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	_, err = db.Collection(USERS_GOOGLE_COLLECTION).InsertOne(context.Background(), user)

	if err != nil {
		return nil, err
	}

	return user, nil
}
