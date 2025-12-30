package middleware

import (
	"net/http"

	"homework-manager/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const UserRoleKey = "user_role"
const UserNameKey = "user_name"

func AuthRequired(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(UserIDKey)

		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		user, err := authService.GetUserByID(userID.(uint))
		if err != nil {
			session.Clear()
			session.Save()
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set(UserIDKey, user.ID)
		c.Set(UserRoleKey, user.Role)
		c.Set(UserNameKey, user.Name)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(UserRoleKey)
		if !exists || role != "admin" {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"title":   "アクセス拒否",
				"message": "この操作には管理者権限が必要です。",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func GuestOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(UserIDKey)

		if userID != nil {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

func InjectUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(UserIDKey)

		if userID != nil {
			c.Set(UserIDKey, userID.(uint))
			c.Set(UserRoleKey, session.Get(UserRoleKey))
			c.Set(UserNameKey, session.Get(UserNameKey))
		}

		c.Next()
	}
}

type APIKeyValidator interface {
	ValidateAPIKey(key string) (uint, error)
}

func APIKeyAuth(validator APIKeyValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use: Bearer <api_key>"})
			c.Abort()
			return
		}

		apiKey := authHeader[len(bearerPrefix):]

		userID, err := validator.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

