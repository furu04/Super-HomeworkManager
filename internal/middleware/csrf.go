package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	csrfTokenKey     = "csrf_token"
	csrfTokenFormKey = "_csrf"
	csrfTokenHeader  = "X-CSRF-Token"
)

type CSRFConfig struct {
	Secret string
}

func generateCSRFToken(secret string) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(randomBytes)
	signature := h.Sum(nil)

	token := append(randomBytes, signature...)
	return base64.URLEncoding.EncodeToString(token), nil
}

func validateCSRFToken(token, secret string) bool {
	if token == "" {
		return false
	}

	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	if len(decoded) != 64 {
		return false
	}

	randomBytes := decoded[:32]
	providedSignature := decoded[32:]

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(randomBytes)
	expectedSignature := h.Sum(nil)

	return hmac.Equal(providedSignature, expectedSignature)
}

func CSRF(config CSRFConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		csrfToken, ok := session.Get(csrfTokenKey).(string)
		if !ok || csrfToken == "" || !validateCSRFToken(csrfToken, config.Secret) {
			newToken, err := generateCSRFToken(config.Secret)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			csrfToken = newToken
			session.Set(csrfTokenKey, csrfToken)
			session.Save()
		}

		c.Set(csrfTokenKey, csrfToken)

		method := strings.ToUpper(c.Request.Method)
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		submittedToken := c.PostForm(csrfTokenFormKey)
		if submittedToken == "" {
			submittedToken = c.GetHeader(csrfTokenHeader)
		}

		sessionToken := session.Get(csrfTokenKey)
		if sessionToken == nil || submittedToken != sessionToken.(string) {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"title":   "CSRFエラー",
				"message": "不正なリクエストです。ページを再読み込みしてください。",
			})
			c.Abort()
			return
		}

		c.Next()

		newToken, err := generateCSRFToken(config.Secret)
		if err == nil {
			session.Set(csrfTokenKey, newToken)
			session.Save()
		}
	}
}

func CSRFField(c *gin.Context) template.HTML {
	token, exists := c.Get(csrfTokenKey)
	if !exists {
		return ""
	}
	return template.HTML(`<input type="hidden" name="` + csrfTokenFormKey + `" value="` + token.(string) + `">`)
}
