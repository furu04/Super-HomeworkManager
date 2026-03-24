package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/dchest/captcha"
)

type CaptchaService struct {
	captchaType        string
	turnstileSecretKey string
}

func NewCaptchaService(captchaType, turnstileSecretKey string) *CaptchaService {
	return &CaptchaService{
		captchaType:        captchaType,
		turnstileSecretKey: turnstileSecretKey,
	}
}

func (s *CaptchaService) NewImageCaptcha() string {
	return captcha.New()
}
func (s *CaptchaService) VerifyImageCaptcha(id, answer string) bool {
	return captcha.VerifyString(id, answer)
}

type turnstileResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

func (s *CaptchaService) VerifyTurnstile(token, remoteIP string) (bool, error) {
	if token == "" {
		return false, nil
	}

	form := url.Values{}
	form.Set("secret", s.turnstileSecretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	resp, err := http.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return false, fmt.Errorf("turnstile request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("reading turnstile response: %w", err)
	}

	var result turnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("parsing turnstile response: %w", err)
	}

	return result.Success, nil
}

func (s *CaptchaService) Type() string {
	return s.captchaType
}
