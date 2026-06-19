package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env              string
	Port             string
	DatabaseURL      string
	JWTSecret        string
	CORSAllowed      []string
	BossEmail        string
	LocalRetention   time.Duration
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	SMTPFromEmail    string
	SMTPFromName     string
	MailWorkerEnabled bool
}

func Load() Config {
	cfg := Config{
		Env:               getEnv("APP_ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		CORSAllowed:       splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:8080")),
		BossEmail:         os.Getenv("BOSS_EMAIL"),
		LocalRetention:    time.Duration(getEnvInt("LOCAL_RETENTION_DAYS", 60)) * 24 * time.Hour,
		AccessTokenTTL:    time.Duration(getEnvInt("ACCESS_TOKEN_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL:   time.Duration(getEnvInt("REFRESH_TOKEN_DAYS", 30)) * 24 * time.Hour,
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          getEnvInt("SMTP_PORT", 587),
		SMTPUsername:      os.Getenv("SMTP_USERNAME"),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
		SMTPFromEmail:     os.Getenv("SMTP_FROM_EMAIL"),
		SMTPFromName:      getEnv("SMTP_FROM_NAME", "AntenaManager PRO"),
		MailWorkerEnabled: getEnvBool("MAIL_WORKER_ENABLED", false),
	}

	if cfg.DatabaseURL == "" {
		log.Println("warning: DATABASE_URL is empty; database endpoints will fail until configured")
	}
	if cfg.JWTSecret == "" {
		log.Println("warning: JWT_SECRET is empty; set a strong random secret before production")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
