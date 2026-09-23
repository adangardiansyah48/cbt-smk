package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseService   string
	JWTSecret         string
	FrontendOrigin    string
}

func Load() *Config {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	origin := os.Getenv("FRONTEND_ORIGIN")
	if origin == "" {
		origin = "http://localhost:5173"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	return &Config{
		Port:            port,
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		SupabaseURL:     os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey: os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseService: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		JWTSecret:       secret,
		FrontendOrigin:  origin,
	}
}
