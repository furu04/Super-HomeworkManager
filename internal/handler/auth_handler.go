package handler

import (
	"net/http"

	"homework-manager/internal/middleware"
	"homework-manager/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	RenderHTML(c, http.StatusOK, "login.html", gin.H{
		"title": "ログイン",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	user, err := h.authService.Login(email, password)
	if err != nil {
		RenderHTML(c, http.StatusOK, "login.html", gin.H{
			"title": "ログイン",
			"error": "メールアドレスまたはパスワードが正しくありません",
			"email": email,
		})
		return
	}

	session := sessions.Default(c)
	session.Set(middleware.UserIDKey, user.ID)
	session.Set(middleware.UserRoleKey, user.Role)
	session.Set(middleware.UserNameKey, user.Name)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	RenderHTML(c, http.StatusOK, "register.html", gin.H{
		"title": "新規登録",
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	passwordConfirm := c.PostForm("password_confirm")
	name := c.PostForm("name")

	if password != passwordConfirm {
		RenderHTML(c, http.StatusOK, "register.html", gin.H{
			"title": "新規登録",
			"error": "パスワードが一致しません",
			"email": email,
			"name":  name,
		})
		return
	}

	if len(password) < 8 {
		RenderHTML(c, http.StatusOK, "register.html", gin.H{
			"title": "新規登録",
			"error": "パスワードは8文字以上で入力してください",
			"email": email,
			"name":  name,
		})
		return
	}

	user, err := h.authService.Register(email, password, name)
	if err != nil {
		errorMsg := "登録に失敗しました"
		if err == service.ErrEmailAlreadyExists {
			errorMsg = "このメールアドレスは既に使用されています"
		}
		RenderHTML(c, http.StatusOK, "register.html", gin.H{
			"title": "新規登録",
			"error": errorMsg,
			"email": email,
			"name":  name,
		})
		return
	}

	session := sessions.Default(c)
	session.Set(middleware.UserIDKey, user.ID)
	session.Set(middleware.UserRoleKey, user.Role)
	session.Set(middleware.UserNameKey, user.Name)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.Redirect(http.StatusFound, "/login")
}
