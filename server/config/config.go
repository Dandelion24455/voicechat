package config

import "os"

type Config struct {
	Port            string
	Domain          string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPass          string
	DBName          string
	RedisAddr       string
	JWTSecret       string
	MediasoupWSHost string
	MediasoupWSPort string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Domain:          getEnv("DOMAIN", ""),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "voicechat"),
		DBPass:          getEnv("DB_PASS", "voicechat"),
		DBName:          getEnv("DB_NAME", "voicechat"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:       getEnv("JWT_SECRET", "jwt-secret-change-me"),
		MediasoupWSHost: getEnv("MEDIASOUP_WS_HOST", "localhost"),
		MediasoupWSPort: getEnv("MEDIASOUP_WS_PORT", "3000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
