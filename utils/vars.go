package utils

import (
	"fmt"
	"os"
)

const (
	SHEET_BASE_URL     string = "SHEET_BASE_URL"
	XLSX_FILE_PATH     string = "XLSX_FILE_PATH"
	KEY_COLUMN         string = "KEY_COLUMN"
	VALUE_COLUMN       string = "VALUE_COLUMN"
	TELEGRAM_BOT_TOKEN string = "TELEGRAM_BOT_TOKEN"
	SERVICE_HOST       string = "SERVICE_HOST"
	VALKEY_HOST        string = "VALKEY_HOST"
)

func AssertEnvVars() error {
	if len(os.Getenv(SHEET_BASE_URL)) == 0 {
		return fmt.Errorf("Missing env var %s", SHEET_BASE_URL)
	}
	if len(os.Getenv(XLSX_FILE_PATH)) == 0 {
		return fmt.Errorf("Missing env var %s", XLSX_FILE_PATH)
	}
	if len(os.Getenv(KEY_COLUMN)) == 0 {
		return fmt.Errorf("Missing env var %s", KEY_COLUMN)
	}
	if len(os.Getenv(VALUE_COLUMN)) == 0 {
		return fmt.Errorf("Missing env var %s", VALUE_COLUMN)
	}
	if len(os.Getenv(TELEGRAM_BOT_TOKEN)) == 0 {
		return fmt.Errorf("Missing env var %s", TELEGRAM_BOT_TOKEN)
	}
	if len(os.Getenv(SERVICE_HOST)) == 0 {
		return fmt.Errorf("Missing env var %s", SERVICE_HOST)
	}
	if len(os.Getenv(VALKEY_HOST)) == 0 {
		return fmt.Errorf("Missing env var %s", VALKEY_HOST)
	}

	return nil
}
