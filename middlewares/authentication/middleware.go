package authentication

import (
	"net/http"
	"os"
	"time"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users/tokens"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")

		if token == "" {

			token = c.GetHeader("Authorization")

			if token == "" && len(token) < 7 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})				
			}

			token = token[7:]
			
			if token == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
			}
		}

		if token == "" && err != nil {
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error parsing the token" + err.Error() + "\n Toke:" + token}) 
		}

		claim = t.Claims.(jwt.MapClaims)

		userInfo, err := tokens.GetUserInfo(claim)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}
		
		if time.Now().UTC().Unix() > int64(claim["exp"].(float64)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})	
		}

		c.Set("userInfo", userInfo)		
		c.Next()
	}
}