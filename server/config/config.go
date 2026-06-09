package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	Domain            string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPass            string
	DBName            string
	RedisAddr         string
	JWTSecret         string
	JWTExpiryHours    int
	AllowedOrigins    []string
	MediasoupWSHost   string
	MediasoupWSPort   string
	MediasoupWSSecure bool
}

func Load() *Config {
	origins := strings.Split(getEnv("ALLOWED_ORIGINS", ""), ",")
	if len(origins) == 1 && origins[0] == "" {
		origins = nil
	}

	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "1"))
	if jwtExpiry <= 0 {
		jwtExpiry = 1
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		Domain:            getEnv("DOMAIN", ""),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "voicechat"),
		DBPass:            getEnv("DB_PASS", "voicechat"),
		DBName:            getEnv("DB_NAME", "voicechat"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:         getEnv("JWT_SECRET", "jwt-secret-change-me"),
		JWTExpiryHours:    jwtExpiry,
		AllowedOrigins:    origins,
		MediasoupWSHost:   getEnv("MEDIASOUP_WS_HOST", "localhost"),
		MediasoupWSPort:   getEnv("MEDIASOUP_WS_PORT", "3000"),
		MediasoupWSSecure: getEnv("MEDIASOUP_WS_SECURE", "") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
