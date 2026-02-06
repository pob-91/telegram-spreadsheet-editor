package model

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

const (
	INPUT_TYPE_TELEGRAM string = "telegram"
)

const (
	SOURCE_TYPE_NEXTCLOUD string = "nextcloud"
)

type Input interface {
	GetType() string
}

type BaseInput struct {
	Type string `yaml:"type"`
}

func (b BaseInput) GetType() string {
	return b.Type
}

type TelegramInput struct {
	BaseInput `yaml:",inline"`
	UserId    int64  `yaml:"userId"`
	TokenEnv  string `yaml:"tokenEnv"`
}

type SpreadsheetSource interface {
	GetType() string
}

type BaseSpreadsheetSource struct {
	Type string `yaml:"type"`
}

func (b BaseSpreadsheetSource) GetType() string {
	return b.Type
}

type NextcloudSpreadsheetSource struct {
	BaseSpreadsheetSource `yaml:",inline"`
	User                  string `yaml:"user"`
	PasswordEnv           string `yaml:"passwordEnv"`
	BaseUrl               string `yaml:"baseUrl"`
	FilePath              string `yaml:"filePath"`
	CostNameColumn        string `yaml:"costNameColumn"`
	CostValueColumn       string `yaml:"costValueColumn"`
	EarningNameColumn     string `yaml:"earningNameColumn"`
	EarningsValueColumn   string `yaml:"earningsValueColumn"`
	StartRow              int    `yaml:"startRow"`
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

	return &cfg, nil
}

// Unmarshalling

func (u *User) UnmarshalYAML(node *yaml.Node) error {
	// capture the raw data
	type rawUser struct {
		Name              string      `yaml:"name"`
		Inputs            []yaml.Node `yaml:"inputs"`
		SpreadsheetSource yaml.Node   `yaml:"spreadsheetSource"`
	}

	var raw rawUser
	if err := node.Decode(&raw); err != nil {
		return err
	}

	u.Name = raw.Name

	for _, inputNode := range raw.Inputs {
		var base BaseInput
		if err := inputNode.Decode(&base); err != nil {
			return fmt.Errorf("Failed to decode input node: %w", err)
		}

		var input Input
		switch base.Type {
		case INPUT_TYPE_TELEGRAM:
			var t TelegramInput
			if err := inputNode.Decode(&t); err != nil {
				return fmt.Errorf("Failed to decode telegram input node: %w", err)
			}
			input = &t
		default:
			return fmt.Errorf("unknown input type: %s", base.Type)
		}

		u.Inputs = append(u.Inputs, input)
	}

	var baseSource BaseSpreadsheetSource
	if err := raw.SpreadsheetSource.Decode(&baseSource); err != nil {
		return fmt.Errorf("failed to decode spreadsheet source type: %w", err)
	}

	switch baseSource.Type {
	case SOURCE_TYPE_NEXTCLOUD:
		var ns NextcloudSpreadsheetSource
		if err := raw.SpreadsheetSource.Decode(&ns); err != nil {
			return fmt.Errorf("failed to decode nextcloud source: %w", err)
		}
		u.SpreadsheetSource = &ns
	default:
		return fmt.Errorf("unknown spreadsheet source type: %s", baseSource.Type)
	}

	return nil
}
