package handler

import (
	"net/http"
	"strconv"

	"homework-manager/internal/middleware"
	"homework-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService  *service.AdminService
	apiKeyService *service.APIKeyService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		adminService:  service.NewAdminService(),
		apiKeyService: service.NewAPIKeyService(),
	}
}

func (h *AdminHandler) getUserID(c *gin.Context) uint {
	userID, _ := c.Get(middleware.UserIDKey)
	return userID.(uint)
}

func (h *AdminHandler) Index(c *gin.Context) {
	users, _ := h.adminService.GetAllUsers()
	currentUserID := h.getUserID(c)

	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "admin/users.html", gin.H{
		"title":         "ユーザー管理",
		"users":         users,
		"currentUserID": currentUserID,
		"isAdmin":       true,
		"userName":      name,
	})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminID := h.getUserID(c)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なユーザーID"})
		return
	}

	err = h.adminService.DeleteUser(adminID, uint(targetID))
	if err != nil {
		users, _ := h.adminService.GetAllUsers()
		name, _ := c.Get(middleware.UserNameKey)

		RenderHTML(c, http.StatusOK, "admin/users.html", gin.H{
			"title":         "ユーザー管理",
			"users":         users,
			"currentUserID": adminID,
			"error":         err.Error(),
			"isAdmin":       true,
			"userName":      name,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

func (h *AdminHandler) ChangeRole(c *gin.Context) {
	adminID := h.getUserID(c)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なユーザーID"})
		return
	}
	newRole := c.PostForm("role")

	err = h.adminService.ChangeRole(adminID, uint(targetID), newRole)
	if err != nil {
		users, _ := h.adminService.GetAllUsers()
		name, _ := c.Get(middleware.UserNameKey)

		RenderHTML(c, http.StatusOK, "admin/users.html", gin.H{
			"title":         "ユーザー管理",
			"users":         users,
			"currentUserID": adminID,
			"error":         err.Error(),
			"isAdmin":       true,
			"userName":      name,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

func (h *AdminHandler) APIKeys(c *gin.Context) {
	keys, _ := h.apiKeyService.GetAllAPIKeys()
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "admin/api_keys.html", gin.H{
		"title":    "APIキー管理",
		"apiKeys":  keys,
		"isAdmin":  true,
		"userName": name,
	})
}

func (h *AdminHandler) CreateAPIKey(c *gin.Context) {
	userID := h.getUserID(c)
	keyName := c.PostForm("name")

	plainKey, _, err := h.apiKeyService.CreateAPIKey(userID, keyName)
	keys, _ := h.apiKeyService.GetAllAPIKeys()
	name, _ := c.Get(middleware.UserNameKey)

	if err != nil {
		RenderHTML(c, http.StatusOK, "admin/api_keys.html", gin.H{
			"title":    "APIキー管理",
			"apiKeys":  keys,
			"error":    err.Error(),
			"isAdmin":  true,
			"userName": name,
		})
		return
	}

	RenderHTML(c, http.StatusOK, "admin/api_keys.html", gin.H{
		"title":      "APIキー管理",
		"apiKeys":    keys,
		"newKey":     plainKey,
		"newKeyName": keyName,
		"isAdmin":    true,
		"userName":   name,
	})
}

func (h *AdminHandler) DeleteAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なAPIキーID"})
		return
	}

	err = h.apiKeyService.DeleteAPIKey(uint(id))
	if err != nil {
		keys, _ := h.apiKeyService.GetAllAPIKeys()
		name, _ := c.Get(middleware.UserNameKey)

		RenderHTML(c, http.StatusOK, "admin/api_keys.html", gin.H{
			"title":    "APIキー管理",
			"apiKeys":  keys,
			"error":    err.Error(),
			"isAdmin":  true,
			"userName": name,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/api-keys")
}

