package handler

import (
	"net/http"

	"homework-manager/internal/middleware"
	"homework-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	authService *service.AuthService
}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{
		authService: service.NewAuthService(),
	}
}

func (h *ProfileHandler) getUserID(c *gin.Context) uint {
	userID, _ := c.Get(middleware.UserIDKey)
	return userID.(uint)
}

func (h *ProfileHandler) Show(c *gin.Context) {
	userID := h.getUserID(c)
	user, _ := h.authService.GetUserByID(userID)

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":    "プロフィール",
		"user":     user,
		"isAdmin":  role == "admin",
		"userName": name,
	})
}

func (h *ProfileHandler) Update(c *gin.Context) {
	userID := h.getUserID(c)
	name := c.PostForm("name")

	err := h.authService.UpdateProfile(userID, name)

	role, _ := c.Get(middleware.UserRoleKey)
	user, _ := h.authService.GetUserByID(userID)

	if err != nil {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":    "プロフィール",
			"user":     user,
			"error":    "プロフィールの更新に失敗しました",
			"isAdmin":  role == "admin",
			"userName": name,
		})
		return
	}

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":    "プロフィール",
		"user":     user,
		"success":  "プロフィールを更新しました",
		"isAdmin":  role == "admin",
		"userName": user.Name,
	})
}

func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	userID := h.getUserID(c)
	oldPassword := c.PostForm("old_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)
	user, _ := h.authService.GetUserByID(userID)

	if newPassword != confirmPassword {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":         "プロフィール",
			"user":          user,
			"passwordError": "新しいパスワードが一致しません",
			"isAdmin":       role == "admin",
			"userName":      name,
		})
		return
	}

	if len(newPassword) < 8 {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":         "プロフィール",
			"user":          user,
			"passwordError": "パスワードは8文字以上で入力してください",
			"isAdmin":       role == "admin",
			"userName":      name,
		})
		return
	}

	err := h.authService.ChangePassword(userID, oldPassword, newPassword)
	if err != nil {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":         "プロフィール",
			"user":          user,
			"passwordError": "現在のパスワードが正しくありません",
			"isAdmin":       role == "admin",
			"userName":      name,
		})
		return
	}

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":           "プロフィール",
		"user":            user,
		"passwordSuccess": "パスワードを変更しました",
		"isAdmin":         role == "admin",
		"userName":        name,
	})
}
