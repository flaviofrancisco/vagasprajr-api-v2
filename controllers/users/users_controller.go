package users

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users/tokens"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/emails"
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
	
	userToken := tokens.UserToken{
		Id: currentUser.Id,
	}

	err = userToken.SetToken(userInfo)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokenExpirationDate := userToken.ExpirationDate.Time().UTC()

	userToken.SetTokenCookie(context)		
	err = tokens.SaveRefreshToken(userInfo)

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

func RefreshToken(context *gin.Context) {
	
	userInfo, err := tokens.ExtractUserInfoForTokenRefresh(context)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user_refresh_token := tokens.UserToken{
		UserId: userInfo.Id,
	}	

	err = user_refresh_token.SetRefreshToken()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if user_refresh_token.Token == "" {
		context.JSON(http.StatusUnauthorized, users.AuthResponse{})
		return
	}

	fmt.Println("Refresh token expiration date (UTC):", user_refresh_token.ExpirationDate.Time())

	// Checks if the user_refresh_token ExpirationDate is in the past
	if user_refresh_token.ExpirationDate.Time().UTC().Before(time.Now().UTC()) {
		tokens.DeleteTokenCookie(context)
		context.JSON(http.StatusUnauthorized, users.AuthResponse{})
		return
	}

	userToken := tokens.UserToken{
		Id: user_refresh_token.Id,
		UserId: user_refresh_token.UserId,
	}
	
	err = userToken.SetToken(userInfo)	

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userToken.SetTokenCookie(context)
	tokenExpirationDate := userToken.ExpirationDate.Time().UTC()

	result := users.AuthResponse {
		AccessToken: userToken.Token,
		Success: true,
		UserInfo: userInfo,
		ExpirationDate: primitive.NewDateTimeFromTime(tokenExpirationDate),
	}	

	context.JSON(http.StatusOK, result)
}

func LogOut(context *gin.Context) {			
	tokens.DeleteTokenCookie(context)
	context.JSON(http.StatusOK, gin.H{"message": "Logout realizado com sucesso"})
}

func ResetPassword(context *gin.Context) {
	var request ResetPasswordRequestBody
	context.BindJSON(&request)

	if request.Token == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Token não informado"})
	}

	if request.Password == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Senha não informada"})
	}

	user, err := users.GetUserByValidationToken(request.Token)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.Email == "" {
		context.JSON(http.StatusNotFound, gin.H{"error": "Token inválido"})
		return
	}

	err = user.UpdatePassword(request.Password)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = user.ResetValidationToken()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"success": true})
}

func VerifyRessetToken(context *gin.Context) {
	
	var request VerifyResestTokenRequest
	context.BindJSON(&request)

	if request.Token == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Token não informado"})
	}

	user, err := users.GetUserByValidationToken(request.Token)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.Email == "" {
		context.JSON(http.StatusNotFound, gin.H{"error": "Token inválido"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"success": true})
}

func RequestPasswordReset(context *gin.Context) {
	
	var request ResetPasswordRequest

	context.BindJSON(&request)

	if request.Email == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "E-mail não informado"})
	}

	user, err := users.GetUserByEmail(request.Email)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.Email == "" {
		context.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	user.ValidationToken = tokens.GenerateValidationToken()	
	err = users.UpdateValidationToken(user)

	if (err != nil) {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	SendRecoveryEmail(user)
}

func SendRecoveryEmail(user users.User) {
	go emails.SendEmail("", []string{user.Email}, "Alteração de senha", "Olá, "+user.FirstName+" "+user.LastName+".\n\n"+"Para alterar sua senha, acesse o link abaixo:\n\n"+os.Getenv("BASE_UI_HOST")+"/alterar-senha?token="+user.ValidationToken+"\n\n"+"Atenciosamente,\n\n"+"Equipe @vagasprajr")
}

func GetUserProfile(context *gin.Context) {
	currentUser, context_error := context.Get("userInfo")

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}
	
	
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

func GetUsers(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"message": "GetUsers"})
}