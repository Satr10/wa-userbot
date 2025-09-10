package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OwnerID string
}

func LoadConfig(filePath string) (Config, error) {
	err := godotenv.Load(filePath)
	if err != nil {
		return Config{}, err
	}

	return Config{
		OwnerID: os.Getenv("OWNER_ID"),
	}, nil

}
