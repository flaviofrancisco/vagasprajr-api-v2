package tokens

import (
	"net/http"
	"time"
	
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/flaviofrancisco/vagasprajr-api-v2/services/authentication"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetRefreshToken(context *gin.Context) {

	token := authentication.Token{}

	userInfo, err := authentication.ValidateToken(context)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userId, err := primitive.ObjectIDFromHex(userInfo.Id)

	if (err != nil) {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user_token := users.UserToken{
		UserId: userId,
	}

	err = user_token.SetRefreshToken()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userInfoFromRefreshToken, err := authentication.ValidateStringToken(user_token.Token)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Checks if token is expired
	if user_token.ExpirationDate.Time().Before(time.Now()) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
		return
	}	

	if userInfo.Email != userInfoFromRefreshToken.Email {
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