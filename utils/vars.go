package utils

import (
	"fmt"
	"os"
)

const (
	CONFIG_PATH_KEY string = "CONFIG_PATH"
	VALKEY_HOST_KEY string = "VALKEY_HOST"
	EVIRONMENT_KEY  string = "ENVIRONMENT"
)

func AssertEnvVars() error {
	if len(os.Getenv(CONFIG_PATH_KEY)) == 0 {
		return fmt.Errorf("Missing env var %s", CONFIG_PATH_KEY)
	}
	if len(os.Getenv(VALKEY_HOST_KEY)) == 0 {
		return fmt.Errorf("Missing env var %s", VALKEY_HOST_KEY)
	}

	return nil
}

func IsDevelopment() bool {
	return os.Getenv(EVIRONMENT_KEY) == "development"
}
