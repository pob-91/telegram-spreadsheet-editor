package model

import (
	"os"

	"go.yaml.in/yaml/v3"
)

type Input struct {
	Type string `yaml:"type"`
}

type TelegramInput struct {
	Input    `yaml:",inline"`
	UserId   int64  `yaml:"userId"`
	TokenEnv string `yaml:"tokenEnv"`
}

type SpreadsheetSource struct {
	Type string `yaml:"type"`
}

type NextcloudSpreadsheetSource struct {
	SpreadsheetSource   `yaml:",inline"`
	User                string `yaml:"user"`
	PasswordEnv         string `yaml:"passwordEnv"`
	BaseUrl             string `yaml:"baseUrl"`
	FilePath            string `yaml:"filePath"`
	CostNameColumn      string `yaml:"costNameColumn"`
	CostValueColumn     string `yaml:"costValueColumn"`
	EarningNameColumn   string `yaml:"earningNameColumn"`
	EarningsValueColumn string `yaml:"earningsValueColumn"`
	StartRow            int    `yaml:"startRow"`
}

type User struct {
	Name              string            `yaml:"name"`
	Inputs            []Input           `yaml:"inputs"`
	SpreadsheetSource SpreadsheetSource `yaml:"spreadsheetSource"`
}

type Config struct {
	Users []User `yaml:"users"`
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// TODO: We need to check the type of spreadsheet source and each input and then serialise them directly into the correct type

	return nil, nil
}
