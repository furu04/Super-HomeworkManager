package handler

import (
	"net/http"

	"homework-manager/internal/middleware"
	"homework-manager/internal/models"
	"homework-manager/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	authService         *service.AuthService
	totpService         *service.TOTPService
	notificationService *service.NotificationService
	appName             string
}

func NewProfileHandler(notificationService *service.NotificationService) *ProfileHandler {
	return &ProfileHandler{
		authService:         service.NewAuthService(),
		totpService:         service.NewTOTPService(),
		notificationService: notificationService,
		appName:             "Super-HomeworkManager",
	}
}

func (h *ProfileHandler) getUserID(c *gin.Context) uint {
	userID, _ := c.Get(middleware.UserIDKey)
	return userID.(uint)
}

func (h *ProfileHandler) Show(c *gin.Context) {
	userID := h.getUserID(c)
	user, _ := h.authService.GetUserByID(userID)
	notifySettings, _ := h.notificationService.GetUserSettings(userID)

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":          "プロフィール",
		"user":           user,
		"isAdmin":        role == "admin",
		"userName":       name,
		"notifySettings": notifySettings,
	})
}

func (h *ProfileHandler) Update(c *gin.Context) {
	userID := h.getUserID(c)
	name := c.PostForm("name")

	err := h.authService.UpdateProfile(userID, name)

	role, _ := c.Get(middleware.UserRoleKey)
	user, _ := h.authService.GetUserByID(userID)
	notifySettings, _ := h.notificationService.GetUserSettings(userID)

	if err != nil {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":          "プロフィール",
			"user":           user,
			"error":          "プロフィールの更新に失敗しました",
			"isAdmin":        role == "admin",
			"userName":       name,
			"notifySettings": notifySettings,
		})
		return
	}

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":          "プロフィール",
		"user":           user,
		"success":        "プロフィールを更新しました",
		"isAdmin":        role == "admin",
		"userName":       user.Name,
		"notifySettings": notifySettings,
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
	notifySettings, _ := h.notificationService.GetUserSettings(userID)

	if newPassword != confirmPassword {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":          "プロフィール",
			"user":           user,
			"passwordError":  "新しいパスワードが一致しません",
			"isAdmin":        role == "admin",
			"userName":       name,
			"notifySettings": notifySettings,
		})
		return
	}

	if len(newPassword) < 8 {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":          "プロフィール",
			"user":           user,
			"passwordError":  "パスワードは8文字以上で入力してください",
			"isAdmin":        role == "admin",
			"userName":       name,
			"notifySettings": notifySettings,
		})
		return
	}

	err := h.authService.ChangePassword(userID, oldPassword, newPassword)
	if err != nil {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":          "プロフィール",
			"user":           user,
			"passwordError":  "現在のパスワードが正しくありません",
			"isAdmin":        role == "admin",
			"userName":       name,
			"notifySettings": notifySettings,
		})
		return
	}

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":           "プロフィール",
		"user":            user,
		"passwordSuccess": "パスワードを変更しました",
		"isAdmin":         role == "admin",
		"userName":        name,
		"notifySettings":  notifySettings,
	})
}

func (h *ProfileHandler) UpdateNotificationSettings(c *gin.Context) {
	userID := h.getUserID(c)
	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)
	user, _ := h.authService.GetUserByID(userID)

	settings := &models.UserNotificationSettings{
		TelegramEnabled: c.PostForm("telegram_enabled") == "on",
		TelegramChatID:  c.PostForm("telegram_chat_id"),
		NotifyOnCreate:  c.PostForm("notify_on_create") == "on",
	}

	err := h.notificationService.UpdateUserSettings(userID, settings)

	notifySettings, _ := h.notificationService.GetUserSettings(userID)

	if err != nil {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":          "プロフィール",
			"user":           user,
			"notifyError":    "通知設定の更新に失敗しました",
			"isAdmin":        role == "admin",
			"userName":       name,
			"notifySettings": notifySettings,
		})
		return
	}

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":          "プロフィール",
		"user":           user,
		"notifySuccess":  "通知設定を更新しました",
		"isAdmin":        role == "admin",
		"userName":       name,
		"notifySettings": notifySettings,
	})
}

const totpPendingSecretKey = "totp_pending_secret"

func (h *ProfileHandler) ShowTOTPSetup(c *gin.Context) {
	userID := h.getUserID(c)
	user, _ := h.authService.GetUserByID(userID)

	setupData, err := h.totpService.GenerateSecret(user.Email, h.appName)
	if err != nil {
		RenderHTML(c, http.StatusOK, "totp_setup.html", gin.H{
			"title": "2段階認証の設定",
			"error": "シークレットの生成に失敗しました",
		})
		return
	}

	session := sessions.Default(c)
	session.Set(totpPendingSecretKey, setupData.Secret)
	session.Save()

	RenderHTML(c, http.StatusOK, "totp_setup.html", gin.H{
		"title":      "2段階認証の設定",
		"secret":     setupData.Secret,
		"qrCode":     setupData.QRCodeB64,
		"otpAuthURL": setupData.OTPAuthURL,
	})
}

func (h *ProfileHandler) EnableTOTP(c *gin.Context) {
	userID := h.getUserID(c)
	user, _ := h.authService.GetUserByID(userID)

	session := sessions.Default(c)
	secret, ok := session.Get(totpPendingSecretKey).(string)
	if !ok || secret == "" {
		c.Redirect(http.StatusFound, "/profile/totp/setup")
		return
	}

	renderSetupError := func(msg string) {
		data := gin.H{
			"title":  "2段階認証の設定",
			"error":  msg,
			"secret": secret,
		}
		if setupData, err := h.totpService.SetupDataFromSecret(secret, user.Email, h.appName); err == nil {
			data["qrCode"] = setupData.QRCodeB64
			data["otpAuthURL"] = setupData.OTPAuthURL
		}
		RenderHTML(c, http.StatusOK, "totp_setup.html", data)
	}

	password := c.PostForm("password")
	if _, err := h.authService.Login(user.Email, password); err != nil {
		renderSetupError("パスワードが正しくありません")
		return
	}

	code := c.PostForm("totp_code")
	if !h.totpService.Validate(secret, code) {
		renderSetupError("認証コードが正しくありません。もう一度試してください")
		return
	}

	if err := h.authService.EnableTOTP(userID, secret); err != nil {
		RenderHTML(c, http.StatusOK, "totp_setup.html", gin.H{
			"title": "2段階認証の設定",
			"error": "2段階認証の有効化に失敗しました",
		})
		return
	}

	session.Delete(totpPendingSecretKey)
	session.Save()

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)
	notifySettings, _ := h.notificationService.GetUserSettings(userID)
	user, _ = h.authService.GetUserByID(userID)

	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":          "プロフィール",
		"user":           user,
		"totpSuccess":    "2段階認証を有効化しました",
		"isAdmin":        role == "admin",
		"userName":       name,
		"notifySettings": notifySettings,
	})
}

func (h *ProfileHandler) DisableTOTP(c *gin.Context) {
	userID := h.getUserID(c)
	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)
	user, _ := h.authService.GetUserByID(userID)
	notifySettings, _ := h.notificationService.GetUserSettings(userID)

	password := c.PostForm("password")
	if _, err := h.authService.Login(user.Email, password); err != nil {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":          "プロフィール",
			"user":           user,
			"totpError":      "パスワードが正しくありません",
			"isAdmin":        role == "admin",
			"userName":       name,
			"notifySettings": notifySettings,
		})
		return
	}

	if err := h.authService.DisableTOTP(userID); err != nil {
		RenderHTML(c, http.StatusOK, "profile.html", gin.H{
			"title":          "プロフィール",
			"user":           user,
			"totpError":      "2段階認証の無効化に失敗しました",
			"isAdmin":        role == "admin",
			"userName":       name,
			"notifySettings": notifySettings,
		})
		return
	}

	user, _ = h.authService.GetUserByID(userID)
	RenderHTML(c, http.StatusOK, "profile.html", gin.H{
		"title":          "プロフィール",
		"user":           user,
		"totpSuccess":    "2段階認証を無効化しました",
		"isAdmin":        role == "admin",
		"userName":       name,
		"notifySettings": notifySettings,
	})
}
