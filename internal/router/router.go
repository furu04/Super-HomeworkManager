package router

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"homework-manager/internal/config"
	"homework-manager/internal/handler"
	"homework-manager/internal/middleware"
	"homework-manager/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006/01/02")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006/01/02 15:04")
		},
		"formatDateInput": func(t time.Time) string {
			return t.Format("2006-01-02T15:04")
		},
		"isOverdue": func(t time.Time, completed bool) bool {
			return !completed && time.Now().After(t)
		},
		"daysUntil": func(t time.Time) int {
			return int(time.Until(t).Hours() / 24)
		},
		"divideFloat": func(a, b int64) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"multiplyFloat": func(a float64, b float64) float64 {
			return a * b
		},
		"recurringLabel":   service.GetRecurrenceTypeLabel,
		"endTypeLabel":     service.GetEndTypeLabel,
		"recurringSummary": service.FormatRecurringSummary,
		"derefInt": func(i *int) int {
			if i == nil {
				return 0
			}
			return *i
		},
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	}
}

func loadTemplates() (*template.Template, error) {
	tmpl := template.New("").Funcs(getFuncMap())

	baseContent, err := os.ReadFile("web/templates/layouts/base.html")
	if err != nil {
		return nil, err
	}
	templateDirs := []struct {
		pattern string
		prefix  string
	}{
		{"web/templates/auth/*.html", ""},
		{"web/templates/pages/*.html", ""},
		{"web/templates/auth/*.html", ""},
		{"web/templates/pages/*.html", ""},
		{"web/templates/assignments/*.html", "assignments/"},
		{"web/templates/recurring/*.html", "recurring/"},
		{"web/templates/admin/*.html", "admin/"},
	}

	for _, dir := range templateDirs {
		files, err := filepath.Glob(dir.pattern)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			name := dir.prefix + filepath.Base(file)
			content, err := os.ReadFile(file)
			if err != nil {
				return nil, err
			}

			reDefine := regexp.MustCompile(`{{\s*define\s+"([^"]+)"\s*}}`)
			reTemplate := regexp.MustCompile(`{{\s*template\s+"([^"]+)"\s*([^}]*)\s*}}`)

			uniqueBase := reDefine.ReplaceAllStringFunc(string(baseContent), func(m string) string {
				match := reDefine.FindStringSubmatch(m)
				blockName := match[1]
				if blockName == "head" || blockName == "scripts" || blockName == "content" || blockName == "base" {
					return strings.Replace(m, blockName, name+"_"+blockName, 1)
				}
				return m
			})

			uniqueBase = reTemplate.ReplaceAllStringFunc(uniqueBase, func(m string) string {
				match := reTemplate.FindStringSubmatch(m)
				blockName := match[1]
				if blockName == "head" || blockName == "scripts" || blockName == "content" || blockName == "base" {
					return strings.Replace(m, blockName, name+"_"+blockName, 1)
				}
				return m
			})
			uniqueContent := reDefine.ReplaceAllStringFunc(string(content), func(m string) string {
				match := reDefine.FindStringSubmatch(m)
				blockName := match[1]
				if blockName == "head" || blockName == "scripts" || blockName == "content" {
					return strings.Replace(m, blockName, name+"_"+blockName, 1)
				}
				return m
			})

			uniqueContent = reTemplate.ReplaceAllStringFunc(uniqueContent, func(m string) string {
				match := reTemplate.FindStringSubmatch(m)
				blockName := match[1]
				if blockName == "base" {
					return strings.Replace(m, blockName, name+"_"+blockName, 1)
				}
				return m
			})

			combined := uniqueBase + "\n" + uniqueContent
			_, err = tmpl.New(name).Parse(combined)
			if err != nil {
				return nil, err
			}
		}
	}

	return tmpl, nil
}

func Setup(cfg *config.Config) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	if len(cfg.TrustedProxies) > 0 {
		r.SetTrustedProxies(cfg.TrustedProxies)
	}
	tmpl, err := loadTemplates()
	if err != nil {
		panic("Failed to load templates: " + err.Error())
	}
	r.SetHTMLTemplate(tmpl)

	r.Static("/static", "web/static")

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   cfg.HTTPS,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("session", store))
	r.Use(middleware.RequestTimer())

	securityConfig := middleware.SecurityConfig{
		HTTPS: cfg.HTTPS,
	}
	r.Use(middleware.SecurityHeaders(securityConfig))
	r.Use(middleware.ForceHTTPS(securityConfig))

	r.Use(middleware.RateLimit(middleware.RateLimitConfig{
		Enabled:  cfg.RateLimitEnabled,
		Requests: cfg.RateLimitRequests,
		Window:   cfg.RateLimitWindow,
	}))

	csrfMiddleware := middleware.CSRF(middleware.CSRFConfig{
		Secret: cfg.CSRFSecret,
	})

	authService := service.NewAuthService()
	apiKeyService := service.NewAPIKeyService()
	notificationService := service.NewNotificationService(cfg.Notification.TelegramBotToken)

	notificationService.StartReminderScheduler()

	authHandler := handler.NewAuthHandler()
	assignmentHandler := handler.NewAssignmentHandler(notificationService)
	adminHandler := handler.NewAdminHandler()
	profileHandler := handler.NewProfileHandler(notificationService)
	apiHandler := handler.NewAPIHandler()
	apiRecurringHandler := handler.NewAPIRecurringHandler()

	guest := r.Group("/")
	guest.Use(middleware.GuestOnly())
	guest.Use(csrfMiddleware)
	{
		guest.GET("/login", authHandler.ShowLogin)
		guest.POST("/login", authHandler.Login)
		if cfg.AllowRegistration {
			guest.GET("/register", authHandler.ShowRegister)
			guest.POST("/register", authHandler.Register)
		} else {
			guest.GET("/register", func(c *gin.Context) {
				c.HTML(http.StatusForbidden, "error.html", gin.H{
					"title":   "登録無効",
					"message": "新規登録は現在受け付けておりません。",
				})
			})
		}
	}

	auth := r.Group("/")
	auth.Use(middleware.AuthRequired(authService))
	auth.Use(csrfMiddleware)
	{
		auth.GET("/", assignmentHandler.Dashboard)
		auth.POST("/logout", authHandler.Logout)

		auth.GET("/assignments", assignmentHandler.Index)
		auth.GET("/assignments/new", assignmentHandler.New)
		auth.POST("/assignments", assignmentHandler.Create)
		auth.GET("/assignments/:id/edit", assignmentHandler.Edit)
		auth.POST("/assignments/:id", assignmentHandler.Update)
		auth.POST("/assignments/:id/toggle", assignmentHandler.Toggle)
		auth.POST("/assignments/:id/delete", assignmentHandler.Delete)

		auth.GET("/statistics", assignmentHandler.Statistics)
		auth.POST("/statistics/archive-subject", assignmentHandler.ArchiveSubject)
		auth.POST("/statistics/unarchive-subject", assignmentHandler.UnarchiveSubject)

		auth.POST("/recurring/:id/stop", assignmentHandler.StopRecurring)
		auth.POST("/recurring/:id/resume", assignmentHandler.ResumeRecurring)
		auth.POST("/recurring/:id/delete", assignmentHandler.DeleteRecurring)
		auth.GET("/recurring", assignmentHandler.ListRecurring)
		auth.GET("/recurring/:id/edit", assignmentHandler.EditRecurring)
		auth.POST("/recurring/:id", assignmentHandler.UpdateRecurring)

		auth.GET("/profile", profileHandler.Show)
		auth.POST("/profile", profileHandler.Update)
		auth.POST("/profile/password", profileHandler.ChangePassword)
		auth.POST("/profile/notifications", profileHandler.UpdateNotificationSettings)

		admin := auth.Group("/admin")
		admin.Use(middleware.AdminRequired())
		{
			admin.GET("/users", adminHandler.Index)
			admin.POST("/users/:id/delete", adminHandler.DeleteUser)
			admin.POST("/users/:id/role", adminHandler.ChangeRole)

			admin.GET("/api-keys", adminHandler.APIKeys)
			admin.POST("/api-keys", adminHandler.CreateAPIKey)
			admin.POST("/api-keys/:id/delete", adminHandler.DeleteAPIKey)
		}
	}

	api := r.Group("/api/v1")
	api.Use(middleware.APIKeyAuth(apiKeyService))
	{
		api.GET("/assignments", apiHandler.ListAssignments)
		api.GET("/assignments/pending", apiHandler.ListPendingAssignments)
		api.GET("/assignments/completed", apiHandler.ListCompletedAssignments)
		api.GET("/assignments/overdue", apiHandler.ListOverdueAssignments)
		api.GET("/assignments/due-today", apiHandler.ListDueTodayAssignments)
		api.GET("/assignments/due-this-week", apiHandler.ListDueThisWeekAssignments)
		api.GET("/assignments/:id", apiHandler.GetAssignment)
		api.POST("/assignments", apiHandler.CreateAssignment)
		api.PUT("/assignments/:id", apiHandler.UpdateAssignment)
		api.DELETE("/assignments/:id", apiHandler.DeleteAssignment)
		api.PATCH("/assignments/:id/toggle", apiHandler.ToggleAssignment)

		api.GET("/statistics", apiHandler.GetStatistics)

		api.GET("/recurring", apiRecurringHandler.ListRecurring)
		api.GET("/recurring/:id", apiRecurringHandler.GetRecurring)
		api.PUT("/recurring/:id", apiRecurringHandler.UpdateRecurring)
		api.DELETE("/recurring/:id", apiRecurringHandler.DeleteRecurring)
	}

	return r
}
