package config

import (
	"log"
	"os"

	"gopkg.in/ini.v1"
)

type DatabaseConfig struct {
	Driver   string // sqlite, mysql, postgres
	Path     string // SQLiteの場合パス
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type Config struct {
	Port              string
	SessionSecret     string
	Debug             bool
	AllowRegistration bool
	HTTPS             bool
	CSRFSecret        string
	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   int
	TrustedProxies    []string
	Database          DatabaseConfig
}

func Load(configPath string) *Config {
	cfg := &Config{
		Port:              "8080",
		SessionSecret:     "",
		Debug:             true,
		AllowRegistration: true,
		HTTPS:             false,
		CSRFSecret:        "",
		RateLimitEnabled:  true,
		RateLimitRequests: 100,
		RateLimitWindow:   60,
		Database: DatabaseConfig{
			Driver:   "sqlite",
			Path:     "homework.db",
			Host:     "localhost",
			Port:     "3306",
			User:     "root",
			Password: "",
			Name:     "homework_manager",
		},
	}

	if configPath == "" {
		configPath = "config.ini"
	}

	if iniFile, err := ini.Load(configPath); err == nil {
		log.Printf("Loading configuration from %s", configPath)

		section := iniFile.Section("server")
		if section.HasKey("port") {
			cfg.Port = section.Key("port").String()
		}
		if section.HasKey("debug") {
			cfg.Debug = section.Key("debug").MustBool(true)
		}

		section = iniFile.Section("database")
		if section.HasKey("driver") {
			cfg.Database.Driver = section.Key("driver").String()
		}
		if section.HasKey("path") {
			cfg.Database.Path = section.Key("path").String()
		}
		if section.HasKey("host") {
			cfg.Database.Host = section.Key("host").String()
		}
		if section.HasKey("port") {
			cfg.Database.Port = section.Key("port").String()
		}
		if section.HasKey("user") {
			cfg.Database.User = section.Key("user").String()
		}
		if section.HasKey("password") {
			cfg.Database.Password = section.Key("password").String()
		}
		if section.HasKey("name") {
			cfg.Database.Name = section.Key("name").String()
		}

		section = iniFile.Section("session")
		if section.HasKey("secret") {
			cfg.SessionSecret = section.Key("secret").String()
		}

		section = iniFile.Section("auth")
		if section.HasKey("allow_registration") {
			cfg.AllowRegistration = section.Key("allow_registration").MustBool(true)
		}

		section = iniFile.Section("security")
		if section.HasKey("https") {
			cfg.HTTPS = section.Key("https").MustBool(false)
		}
		if section.HasKey("csrf_secret") {
			cfg.CSRFSecret = section.Key("csrf_secret").String()
		}
		if section.HasKey("rate_limit_enabled") {
			cfg.RateLimitEnabled = section.Key("rate_limit_enabled").MustBool(true)
		}
		if section.HasKey("rate_limit_requests") {
			cfg.RateLimitRequests = section.Key("rate_limit_requests").MustInt(100)
		}
		if section.HasKey("rate_limit_window") {
			cfg.RateLimitWindow = section.Key("rate_limit_window").MustInt(60)
		}
		if section.HasKey("trusted_proxies") {
			proxies := section.Key("trusted_proxies").String()
			if proxies != "" {
				cfg.TrustedProxies = []string{proxies}
			}
		}
	} else {
		log.Println("config.ini not found, using environment variables or defaults")
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}
	if dbDriver := os.Getenv("DATABASE_DRIVER"); dbDriver != "" {
		cfg.Database.Driver = dbDriver
	}
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if dbHost := os.Getenv("DATABASE_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DATABASE_PORT"); dbPort != "" {
		cfg.Database.Port = dbPort
	}
	if dbUser := os.Getenv("DATABASE_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPassword := os.Getenv("DATABASE_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		cfg.Database.Name = dbName
	}
	if sessionSecret := os.Getenv("SESSION_SECRET"); sessionSecret != "" {
		cfg.SessionSecret = sessionSecret
	}
	if os.Getenv("GIN_MODE") == "release" {
		cfg.Debug = false
	}
	if allowReg := os.Getenv("ALLOW_REGISTRATION"); allowReg != "" {
		cfg.AllowRegistration = allowReg == "true" || allowReg == "1"
	}
	if https := os.Getenv("HTTPS"); https != "" {
		cfg.HTTPS = https == "true" || https == "1"
	}
	if csrfSecret := os.Getenv("CSRF_SECRET"); csrfSecret != "" {
		cfg.CSRFSecret = csrfSecret
	}
	if trustedProxies := os.Getenv("TRUSTED_PROXIES"); trustedProxies != "" {
		cfg.TrustedProxies = []string{trustedProxies}
	}

	if cfg.SessionSecret == "" {
		log.Fatal("FATAL: Session secret is not set. Please set it in config.ini ([session] secret) or via SESSION_SECRET environment variable.")
	}
	if cfg.CSRFSecret == "" {
		log.Fatal("FATAL: CSRF secret is not set. Please set it in config.ini ([security] csrf_secret) or via CSRF_SECRET environment variable.")
	}

	return cfg
}

