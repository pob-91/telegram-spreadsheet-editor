package tests

import (
	"telegram-spreadsheet-editor/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_InitBasicConfig(t *testing.T) {
	// given
	configPath := "../../config.example.yaml"

	// when
	config, err := model.NewConfigFromFile(configPath)

	// then
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Users, 2)

	assert.Equal(t, config.Users[0].Name, "Rob")

	assert.Len(t, config.Users[0].Inputs, 1)
	assert.Equal(t, config.Users[0].Inputs[0].GetType(), "telegram")

	// TOOD Assert types and other properties
}
