package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "auth.claims"

func Middleware(tokens *TokenManager, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		token, err := BearerToken(c.GetHeader("Authorization"))
		if err != nil {
			unauthorized(c, "missing authorization token")
			return
		}
		claims, err := tokens.Verify(token)
		if err != nil {
			unauthorized(c, err.Error())
			return
		}
		if len(allowed) > 0 {
			if _, ok := allowed[claims.Role]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40003, "message": "permission denied", "data": nil})
				return
			}
		}
		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*Claims, bool) {
	value, ok := c.Get(claimsContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*Claims)
	return claims, ok
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    40001,
		"message": message,
		"data":    nil,
	})
}
