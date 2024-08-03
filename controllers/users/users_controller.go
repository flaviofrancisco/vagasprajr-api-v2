package users

import (
	"net/http"
	"strings"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/authentication"
	"github.com/gin-gonic/gin"
)

func CreateUser(context *gin.Context) {

	var body users.User
	context.BindJSON(&body)

	err := users.CreateUser(body)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func Login(context *gin.Context) {	

	var body users.AuthRequestBody
	context.BindJSON(&body)

	var user users.User

	user.Email = body.Email
	user.Password = body.Password

	isAuthenticated, err := user.IsAuthenticated()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !isAuthenticated {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Não foi possível autenticar o usuário"})		
		return
	}
	
	currentUser, err := users.GetUserByEmail(user.Email)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if currentUser.Email == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não encontrado"})
		return
	}

	if !currentUser.IsEmailConfirmed {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não confirmou o e-mail"})
		return
	}

	currentUser.UpdateLastLogin()

	userInfo := users.UserInfo{
		Email: strings.ToLower(currentUser.Email),
		FirstName: currentUser.FirstName,
		LastName: currentUser.LastName,
		Links: currentUser.Links,
		Provider: currentUser.Provider,
		UserName: currentUser.UserName,
		Id: currentUser.Id.Hex(),
	}

	var token authentication.Token
	
	token_string, err := token.GetToken(userInfo, false)
	
	if (err != nil) {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token.Token = token_string

	token.SetTokenCookie(context)	

	err = token.SaveRefreshToken(userInfo)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := users.AuthResponse {
		AccessToken: token.Token,
		Success: true,
		UserInfo: userInfo,
	}	

	context.JSON(http.StatusOK, result)

}