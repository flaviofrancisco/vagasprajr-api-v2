package users

import (
	"net/http"
	"strings"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SignUp (context *gin.Context) {
	
	var body users.User
	err := context.BindJSON(&body)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = users.SignUp(body)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "isRegistered": false})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"isRegistered": true})	
}

func ConfirmEmail(context *gin.Context) {
	
	token := context.Param("token")

	if token == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Token não informado"})
		return
	}

	user, err := users.GetUserByValidationToken(token)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.Email == "" {
		context.JSON(http.StatusNotFound, gin.H{"error": "Token inválido"})
		return
	}
	
	err = user.ConfirmEmail()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"isEmailConfirmed": true})
}

func CreateUser(context *gin.Context) {

	var body users.User
	context.BindJSON(&body)

	err := users.CreateUser(body)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "isRegistered": false})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"isRegistered": true})
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
		Id: currentUser.Id,
	}

	expirationDateUTC := time.Now().UTC().Add(time.Duration(1) * time.Hour)

	userToken := users.UserToken{
		Id: currentUser.Id,
	}

	err = userToken.SetToken(userInfo, expirationDateUTC)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokenExpirationDate := userToken.ExpirationDate.Time().UTC()

	userToken.SetTokenCookie(context)		
	err = users.SaveRefreshToken(userInfo)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := users.AuthResponse {
		AccessToken: userToken.Token,
		Success: true,
		UserInfo: userInfo,
		ExpirationDate: primitive.NewDateTimeFromTime(tokenExpirationDate),		
	}	
	
	context.JSON(http.StatusOK, result)
}

func LogOut(context *gin.Context) {			
	users.DeleteTokenCookie(context)
	context.JSON(http.StatusOK, gin.H{"message": "Logout realizado com sucesso"})
}

func GetUser(context *gin.Context) {
	
	currentUser, context_error := context.Get("userInfo")	

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserInfo)
	
	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := UserProfileResponse {
		Id: user.Id.Hex(),
		FirstName: user.FirstName,
		LastName: user.LastName,
		Email: user.Email,
		UserName: user.UserName,
		AboutMe: user.AboutMe,
		City: user.City,
		State: user.State,
		Links: user.Links,
		IsEmailConfirmed: user.IsEmailConfirmed,
		Roles: user.Roles,
		Experiences: user.Experiences,
		IsPublic: user.IsPublic,
		ProfileViews: user.ProfileViews,
		TechExperiences: user.TechExperiences,
		Educations: user.Educations,
		Certifications: user.Certifications,
		JobPreference: user.JobPreference,
		DiversityInfo: user.DiversityInfo,
		IdiomsInfo: user.IdiomsInfo,
		IsPublicForRecruiter: user.IsPublicForRecruiter,
	}

	context.JSON(http.StatusOK, response)
}

func GetRefreshToken(context *gin.Context) {

	token := users.UserToken{}
	userInfo := users.UserInfo{}

	err := userInfo.GetUserInfoFromContext(context)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user_token := users.UserToken{
		Id: userInfo.Id,
	}	

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = user_token.SetRefreshToken()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userInfo_refresh_token := users.UserInfo{}

	err = userInfo_refresh_token.SetUserInfoFromTokenString(user_token.Token)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Checks if token is expired
	if user_token.ExpirationDate.Time().Before(time.Now()) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
		return
	}	

	if userInfo.Email != userInfo_refresh_token.Email {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	token.ExpirationDate = user_token.ExpirationDate
	token.SetTokenCookie(context)

	result := users.AuthResponse {
		AccessToken: user_token.Token,
		Success: true,
		UserInfo: userInfo,
		ExpirationDate: user_token.ExpirationDate,
	}	

	context.JSON(http.StatusOK, result)
}

func GetUsers(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"message": "GetUsers"})
}