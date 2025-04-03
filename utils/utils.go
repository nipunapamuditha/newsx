package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVariable(key string) (string, error) {

	errr := godotenv.Load("E:\\Personal Projects\\newsx_version_3\\.env")
	if errr != nil {
		log.Printf("Error loading .env file: %v", errr)
		return "", errr
	}

	value := os.Getenv(key)
	if value == "" {
		err := fmt.Errorf("Environment variable %s not correct", key)
		return "", err
	}

	return value, nil
}
