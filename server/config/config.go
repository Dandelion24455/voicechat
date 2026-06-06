package config

import "os"

type Config struct {
	Port       string
	Domain     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	RedisAddr  string
	JWTSecret  string
	LiveKitURL    string
	LiveKitKey    string
	LiveKitSecret string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		Domain:     getEnv("DOMAIN", ""),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "voicechat"),
		DBPass:     getEnv("DB_PASS", "voicechat"),
		DBName:     getEnv("DB_NAME", "voicechat"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "jwt-secret-change-me"),
		LiveKitURL:    getEnv("LIVEKIT_URL", "http://localhost:7880"),
		LiveKitKey:    getEnv("LIVEKIT_API_KEY", getEnv("LIVEKIT_KEY", "devkey")),
		LiveKitSecret: getEnv("LIVEKIT_API_SECRET", getEnv("LIVEKIT_SECRET", "secretsecretsecretsecretsecret12")),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
