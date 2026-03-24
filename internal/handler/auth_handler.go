package handler

import (
	"net/http"

	"homework-manager/internal/config"
	"homework-manager/internal/middleware"
	"homework-manager/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const twoFAPendingKey = "2fa_pending_user_id"

type AuthHandler struct {
	authService    *service.AuthService
	totpService    *service.TOTPService
	captchaService *service.CaptchaService
	captchaCfg     config.CaptchaConfig
}

func NewAuthHandler(captchaCfg config.CaptchaConfig) *AuthHandler {
	captchaSvc := service.NewCaptchaService(captchaCfg.Type, captchaCfg.TurnstileSecretKey)
	return &AuthHandler{
		authService:    service.NewAuthService(),
		totpService:    service.NewTOTPService(),
		captchaService: captchaSvc,
		captchaCfg:     captchaCfg,
	}
}

func (h *AuthHandler) captchaData() gin.H {
	data := gin.H{
		"captchaEnabled": h.captchaCfg.Enabled,
		"captchaType":    h.captchaCfg.Type,
	}
	if h.captchaCfg.Enabled && h.captchaCfg.Type == "turnstile" {
		data["turnstileSiteKey"] = h.captchaCfg.TurnstileSiteKey
	}
	if h.captchaCfg.Enabled && h.captchaCfg.Type == "image" {
		data["captchaID"] = h.captchaService.NewImageCaptcha()
	}
	return data
}

func (h *AuthHandler) verifyCaptcha(c *gin.Context) bool {
	if !h.captchaCfg.Enabled {
		return true
	}
	switch h.captchaCfg.Type {
	case "turnstile":
		token := c.PostForm("cf-turnstile-response")
		ok, err := h.captchaService.VerifyTurnstile(token, c.ClientIP())
		return err == nil && ok
	case "image":
		id := c.PostForm("captcha_id")
		answer := c.PostForm("captcha_answer")
		return h.captchaService.VerifyImageCaptcha(id, answer)
	}
	return true
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	data := gin.H{"title": "ログイン"}
	for k, v := range h.captchaData() {
		data[k] = v
	}
	RenderHTML(c, http.StatusOK, "login.html", data)
}

func (h *AuthHandler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	renderLoginError := func(msg string) {
		data := gin.H{
			"title": "ログイン",
			"error": msg,
			"email": email,
		}
		for k, v := range h.captchaData() {
			data[k] = v
		}
		RenderHTML(c, http.StatusOK, "login.html", data)
	}

	if !h.verifyCaptcha(c) {
		renderLoginError("CAPTCHAの検証に失敗しました。もう一度お試しください")
		return
	}

	user, err := h.authService.Login(email, password)
	if err != nil {
		renderLoginError("メールアドレスまたはパスワードが正しくありません")
		return
	}

	if user.TOTPEnabled {
		session := sessions.Default(c)
		session.Set(twoFAPendingKey, user.ID)
		session.Save()
		c.Redirect(http.StatusFound, "/login/2fa")
		return
	}

	session := sessions.Default(c)
	session.Set(middleware.UserIDKey, user.ID)
	session.Set(middleware.UserRoleKey, user.Role)
	session.Set(middleware.UserNameKey, user.Name)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) ShowLogin2FA(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get(twoFAPendingKey) == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	RenderHTML(c, http.StatusOK, "login_2fa.html", gin.H{
		"title": "2段階認証",
	})
}

func (h *AuthHandler) Login2FA(c *gin.Context) {
	session := sessions.Default(c)
	pendingID := session.Get(twoFAPendingKey)
	if pendingID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID := pendingID.(uint)
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		session.Delete(twoFAPendingKey)
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	code := c.PostForm("totp_code")
	if !h.totpService.Validate(user.TOTPSecret, code) {
		RenderHTML(c, http.StatusOK, "login_2fa.html", gin.H{
			"title": "2段階認証",
			"error": "認証コードが正しくありません",
		})
		return
	}

	session.Delete(twoFAPendingKey)
	session.Set(middleware.UserIDKey, user.ID)
	session.Set(middleware.UserRoleKey, user.Role)
	session.Set(middleware.UserNameKey, user.Name)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	data := gin.H{"title": "新規登録"}
	for k, v := range h.captchaData() {
		data[k] = v
	}
	RenderHTML(c, http.StatusOK, "register.html", data)
}

func (h *AuthHandler) Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	passwordConfirm := c.PostForm("password_confirm")
	name := c.PostForm("name")

	renderRegisterError := func(msg string) {
		data := gin.H{
			"title": "新規登録",
			"error": msg,
			"email": email,
			"name":  name,
		}
		for k, v := range h.captchaData() {
			data[k] = v
		}
		RenderHTML(c, http.StatusOK, "register.html", data)
	}

	if !h.verifyCaptcha(c) {
		renderRegisterError("CAPTCHAの検証に失敗しました。もう一度お試しください")
		return
	}

	if password != passwordConfirm {
		renderRegisterError("パスワードが一致しません")
		return
	}

	if len(password) < 8 {
		renderRegisterError("パスワードは8文字以上で入力してください")
		return
	}

	user, err := h.authService.Register(email, password, name)
	if err != nil {
		errorMsg := "登録に失敗しました"
		if err == service.ErrEmailAlreadyExists {
			errorMsg = "このメールアドレスは既に使用されています"
		}
		renderRegisterError(errorMsg)
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
