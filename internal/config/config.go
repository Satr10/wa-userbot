package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OwnerID      string
	GeminiAPIKey string
	GSBAPIKey    string
}

// TODO:IMPROVE THIS FUNCTION
func LoadConfig(filePath string) (Config, error) {
	err := godotenv.Load(filePath)
	if err != nil {
		return Config{}, err
	}

	return Config{
		OwnerID:      os.Getenv("OWNER_ID"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GSBAPIKey:    os.Getenv("GOOGLE_SAFE_BROWSING_API_KEY"),
	}, nil

}
