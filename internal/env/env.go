package env

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func GetString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("could not find environment value: ", key)
	}
	log.Printf("found value for %s environment value\n", key)

	return value, nil
}

func GetInt(key string) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return 0, fmt.Errorf("could not find environment value: ", key)
	}
	log.Printf("found value for %s environment value\n", key)

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("could not convert %s from string to int: %v", key, err)
	}

	return value, nil
}
