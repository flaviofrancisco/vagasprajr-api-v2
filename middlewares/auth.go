package middlewares

import (
	"net/http"
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/services/authentication"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")

		if token == "" {

			token = c.GetHeader("Authorization")
			token = token[7:]
			
			if token == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
			}
		}

		if token == "" || err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error parsing the token"})
		}

		claim = t.Claims.(jwt.MapClaims)

		userInfo := authentication.GetUserInfoFromClaims(claim)

		if time.Now().Unix() > int64(claim["exp"].(float64)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})	
		}

		c.Set("userInfo", userInfo)		
		c.Next()
	}
}