package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OwnerID      string
	GeminiAPIKey string
	GSBAPIKey    string
	PostgressURI string
}

// TODO:IMPROVE THIS FUNCTION
func LoadConfig() (Config, error) {
	// godotenv.Load() will load a .env file if it exists, but won't fail if it doesn't.
	// This is ideal for environments like Leapcell where vars are set directly.
	godotenv.Load()

	return Config{
		OwnerID:      os.Getenv("OWNER_ID"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GSBAPIKey:    os.Getenv("GOOGLE_SAFE_BROWSING_API_KEY"),
		PostgressURI: os.Getenv("POSTGRES_URI"),
	}, nil

}
