package utils

import (
	"errors"
	"fmt"
	"os"
)

func LoadVar(key string) (string, error) {
	variable := os.Getenv(key)
	if variable == "" {
		return "", errors.New(fmt.Sprintf("%q environment variable not set!", key))
	}

	return variable, nil
}
