package middleware

import (
	"strings"

	"user-center/internal/exception"
	"user-center/internal/util/jwt"
	"github.com/gin-gonic/gin"
)

const (
	CtxUserID   = "userID"
	CtxUsername = "username"
)

func AuthRequired() gin.HandlerFunc {
	mgr := jwt.GetManager()
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		token := ""
		if raw != "" {
			parts := strings.SplitN(raw, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token = parts[1]
			}
		}
		if token == "" {
			token = c.Query("access_token")
		}
		if token == "" {
			exception.Fail(c, exception.NewUnauthorized("missing bearer token"))
			c.Abort()
			return
		}
		claims, ex := mgr.Parse(token)
		if ex != nil {
			exception.Fail(c, ex)
			c.Abort()
			return
		}
		if claims.Type != jwt.TypeAccess {
			exception.Fail(c, exception.NewUnauthorized("invalid token type"))
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "600")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
