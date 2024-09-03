package users

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/middlewares"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/commons"
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/gravatar"
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

	userInfo := users.UserTokenInfo{
		Email: strings.ToLower(currentUser.Email),
		FirstName: currentUser.FirstName,
		LastName: currentUser.LastName,				
		UserName: currentUser.UserName,
		Id: currentUser.Id,
		ProfileImageUrl: currentUser.ProfileImageUrl,
		Roles: currentUser.Roles,
	}
	
	userToken := tokens.UserToken{
		Id: currentUser.Id,
	}

	err = userToken.SetToken(userInfo)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userToken.SetTokenCookie(context)		
	err = tokens.SaveRefreshToken(userInfo)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokenExpirationDate := userToken.ExpirationDate.Time().UTC()

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

	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, users.AuthResponse{})
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

func GetUser(context *gin.Context) {
	
	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}

	userId := context.Param("id")

	if userId == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Id do usuário não informado"})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Id do usuário inválido"})
		return
	}

	user, err := users.GetUserById(objectId)

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
		BookmarkedJobs: user.BookmarkedJobs,
		ProfileImageUrl: user.ProfileImageUrl,
		OAuthImageURL: user.OAuthImageURL,
		GravatarImageUrl: user.GravatarImageUrl,
	}

	context.JSON(http.StatusOK, response)
}

func GetPublicUserProfile(context *gin.Context) {
	
	userName := context.Param("username")

	if userName == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Nome do usuário não informado"})
		return
	}
	
	user, err := users.GetUserByUserName(userName)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !user.IsPublic {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}

	user.IncrementProfileViews()

	response := UserProfileResponse {
		Id: user.Id.Hex(),
		FirstName: user.FirstName,
		LastName: user.LastName,		
		UserName: user.UserName,
		AboutMe: user.AboutMe,
		City: user.City,
		State: user.State,
		Links: user.Links,
		IsEmailConfirmed: user.IsEmailConfirmed,		
		Experiences: user.Experiences,				
		TechExperiences: user.TechExperiences,
		Educations: user.Educations,
		Certifications: user.Certifications,				
		IdiomsInfo: user.IdiomsInfo,
		IsPublic: user.IsPublic,		
		ProfileImageUrl: user.ProfileImageUrl,
	}

	context.JSON(http.StatusOK, response)
}

func DeleteUserAsAdmin(context *gin.Context) {
	
	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}

	userId := context.Param("id")

	if userId == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Id do usuário não informado"})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Id do usuário inválido"})
		return
	}

	err = users.DeleteUser(objectId)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteUser(context *gin.Context) {
	
	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)

	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = users.DeleteUser(user.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"success": true})
}

func GetUserProfile(context *gin.Context) {
	
	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)	

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)
	
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
		BookmarkedJobs: user.BookmarkedJobs,
		ProfileImageUrl: user.ProfileImageUrl,
		OAuthImageURL: user.OAuthImageURL,
		GravatarImageUrl: user.GravatarImageUrl,
	}

	context.JSON(http.StatusOK, response)
}

func IsAuthorized(context *gin.Context) {
	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)

	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.Email == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não encontrado"})
		return
	}

	if userInfo.Id != user.Id {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}
	
	var request AuthorizeRequest
	context.BindJSON(&request)

	isAuthorized := false

	for _, role := range request.Roles {
		if user.IsAuthorized(role) {
			isAuthorized = true
			break
		}
	}

	context.JSON(http.StatusOK, gin.H{"isAuthorized": isAuthorized})
}

func UpdateUserName(context *gin.Context) {
	
	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)	

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)

	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)

	if err != nil {	
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var request UpdateUserNameRequest	
	context.BindJSON(&request)

	if request.UserName == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Nome de usuário não informado"})
		return
	}

	is_valid := users.IsUserNameValid(request.UserName)

	if !is_valid {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Nome de usuário inválido. Somete use letras e números; não use espaço e deve ter no mínimo 3 caracteres. Por favor, escolha outro."})
		return
	}

	already_exists, err := users.UserNameAlreadyExists(request.UserName)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if already_exists {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Nome de usuário já em uso. Por favor, escolha outro."})
		return
	}

	user.Id = userInfo.Id
	user.UserName = request.UserName

	err = user.UpdateUserName()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err = users.GetUserById(userInfo.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"success": true, "user_name": user.UserName, "message": "Nome de usuário atualizado com sucesso"})	
}

func UpdateUser(context *gin.Context) {
	
	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)	

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)
	
	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)
	
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userInfo.Id != user.Id {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}

	var request UpdateUserRequest
	context.BindJSON(&request)

	user.Id = userInfo.Id
	user.FirstName = request.FirstName
	user.LastName = request.LastName
	user.City = request.City
	user.State = request.State
	user.AboutMe = request.AboutMe
	user.Links = request.Links
	user.Experiences = request.Experiences
	user.TechExperiences = request.TechExperiences
	user.IdiomsInfo = request.IdiomsInfo
	user.Certifications = request.Certifications
	user.Educations = request.Educations
	user.IsPublic = request.IsPublic
	user.ProfileImageUrl = request.ProfileImageUrl
	user.OAuthImageURL = request.OAuthImageURL
		
	err = user.Update()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err = users.GetUserById(userInfo.Id)
	
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

func UploadProfilePicture(context *gin.Context) {
	
	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)	

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)
	
	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userInfo.Id != user.Id {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}

	file, err := context.FormFile("file")

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileName := fmt.Sprintf("%s%s", "profile_picture_", file.Filename)

	userIdString := userInfo.Id.Hex()

	err = context.SaveUploadedFile(file, "./uploads/"+ userIdString + "/" + fileName)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"success": true, "profile_picture": fileName, "message": "Foto de perfil atualizada com sucesso"})

	user.ProfileImageUrl = fileName

	err = user.UpdateProfilePicture()

	// if err != nil {
	// 	context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// user, err = users.GetUserById(userInfo.Id)

	// if err != nil {
	// 	context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// context.JSON(http.StatusOK, gin.H{"success": true, "profile_picture": user.ProfilePicture, "message": "Foto de perfil atualizada com sucesso"})
}

func UpdateUserBookmarkedJobs(context *gin.Context) {
	currentUser, context_error := context.Get(middlewares.USER_TOKEN_INFO)	

	if !context_error {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recuperar informações do usuário conectado"})
		return
	}

	userInfo := currentUser.(users.UserTokenInfo)
	
	if userInfo.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, err := users.GetUserById(userInfo.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.Id.IsZero() {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	var request UpdateUserBookmarkRequest
	context.BindJSON(&request)

	user.BookmarkedJobs = request.BookmarkedJobs
	user.UpdateUserBookmarkedJobs()

	user, err = users.GetUserById(userInfo.Id)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, user.BookmarkedJobs)	
}

func GetUsers(context *gin.Context) {

	userRole := context.MustGet("userRole").(string)

	if userRole != "admin" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}

	var request commons.FilterRequest
	context.BindJSON(&request)

	result, err := users.GetUsers(request)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
		
	context.JSON(http.StatusOK, result)	
}

func GetTalents(context *gin.Context) {
	
	var request commons.FilterRequest
	context.BindJSON(&request)

	result, err := users.GetTalents(request, false, false)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
		
	context.JSON(http.StatusOK, result)
}

func GetGravatarUrl(context *gin.Context) {
	
	var request GravatarRequest
	context.BindJSON(&request)

	if request.Email == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "E-mail não informado"})
		return
	}

	gravatar := gravatar.NewGravatarFromEmail(request.Email)

	context.JSON(http.StatusOK, gin.H{"gravatarUrl": gravatar.GetURL()})
}