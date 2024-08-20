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

		userInfo := c.MustGet("userInfo").(users.UserInfo)

		if userInfo.Id.IsZero() {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		}

		userRoles, err := users.GetUserRoles(userInfo.Id)

		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
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

		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})	
		
	}
}