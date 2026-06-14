package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "auth.claims"

func Middleware(tokens *TokenManager, sessions SessionStore, roles ...string) gin.HandlerFunc {
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
		if sessions == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": 50000, "message": "session store unavailable", "data": nil})
			return
		}
		session, err := sessions.Get(c.Request.Context(), claims.SessionID)
		if err != nil {
			if errors.Is(err, ErrSessionNotFound) {
				unauthorized(c, "login session expired")
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": 50000, "message": "session store unavailable", "data": nil})
			return
		}
		if !time.Now().Before(session.ExpiresAt) ||
			session.UserID != claims.UserID ||
			session.Username != claims.Username ||
			session.Role != claims.Role {
			unauthorized(c, "invalid login session")
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
