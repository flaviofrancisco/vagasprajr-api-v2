package authentication

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users/tokens"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	TOKEN_NAME = "vagasprajr_token"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(TOKEN_NAME)

		if token == "" {

			token = c.GetHeader("Authorization")

			if token == "" || len(token) < 7 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})								
				return
			}

			token = token[7:]
			
			if token == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
				return
			}
		}

		if token == "" && err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
			return
		}		
		
		var (
			key   []byte
			t     *jwt.Token
			claim jwt.MapClaims
		)

		key = []byte(os.Getenv("JWT_SECRET"))

        // Check if the token is malformed
        parts := strings.Split(token, ".")
        if len(parts) != 3 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Malformed token"})
            return
        }		

		t, err = jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error parsing the token. " + err.Error() + " Token:" + token}) 
			return
		}

		claim = t.Claims.(jwt.MapClaims)

		userTokenInfo, err := tokens.GetUserInfo(claim)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		
		if time.Now().UTC().Unix() > int64(claim["exp"].(float64)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})	
			return
		}

		c.Set("userTokenInfo", userTokenInfo)		
		c.Next()
	}
}