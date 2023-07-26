package utils

import (
	"log"
	"os"
)

func GetEnv(key string, required bool) string {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatal("Variable: " + key + " is required")
	}
	return value
}

func GetBoolEnv(key string) bool {
	value := os.Getenv(key)
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}
	log.Fatal("Variable: " + key + " must be true or false")
	return false
}
