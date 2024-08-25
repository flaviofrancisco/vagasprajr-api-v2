package authorization

import (
	"github.com/flaviofrancisco/vagasprajr-api-v2/models/users"
	"github.com/gin-gonic/gin"
)

func AuthorizationMiddleware(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if len(roles) == 0 {			
			c.Next()
			return
		}

		userTokenInfo := c.MustGet("userTokenInfo").(users.UserTokenInfo)

		if userTokenInfo.Id.IsZero() {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unable to get user information"})
		}

		userRoles, err := users.GetUserRoles(userTokenInfo.Id)

		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unable to get user roles"})
		}

		for _, role := range roles {
			for _, userRole := range userRoles {
				if role == userRole {
					c.Set("userRole", userRole)	
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(401, gin.H{"error": "User does not have the necessary permissions"})	
		
	}
}