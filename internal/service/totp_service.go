package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/url"

	"github.com/pquerna/otp"
	totplib "github.com/pquerna/otp/totp"
)

type TOTPService struct{}

func NewTOTPService() *TOTPService {
	return &TOTPService{}
}

type TOTPSetupData struct {
	Secret     string
	QRCodeB64  string
	OTPAuthURL string
}

func (s *TOTPService) GenerateSecret(email, issuer string) (*TOTPSetupData, error) {
	key, err := totplib.Generate(totplib.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return nil, err
	}
	return s.buildSetupData(key)
}

func (s *TOTPService) SetupDataFromSecret(secret, email, issuer string) (*TOTPSetupData, error) {
	otpURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		url.PathEscape(issuer),
		url.PathEscape(email),
		url.QueryEscape(secret),
		url.QueryEscape(issuer),
	)
	key, err := otp.NewKeyFromURL(otpURL)
	if err != nil {
		return nil, err
	}
	return s.buildSetupData(key)
}

func (s *TOTPService) buildSetupData(key *otp.Key) (*TOTPSetupData, error) {
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return &TOTPSetupData{
		Secret:     key.Secret(),
		QRCodeB64:  base64.StdEncoding.EncodeToString(buf.Bytes()),
		OTPAuthURL: key.URL(),
	}, nil
}

func (s *TOTPService) Validate(secret, code string) bool {
	return totplib.Validate(code, secret)
}
