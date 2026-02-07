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

	telegramInput, ok := config.Users[0].Inputs[0].(*model.TelegramInput)
	assert.True(t, ok)

	assert.Equal(t, telegramInput.UserId, int64(1234))
	assert.Equal(t, telegramInput.TokenEnv, "ROB_TELEGRAM_TOKEN")

	nextcloudSource, ok := config.Users[0].SpreadsheetSource.(*model.NextcloudSpreadsheetSource)
	assert.True(t, ok)

	assert.Equal(t, nextcloudSource.User, "admin")
	assert.Equal(t, nextcloudSource.PasswordEnv, "ROB_NEXTCLOUD_PASSWORD")
	assert.Equal(t, nextcloudSource.BaseUrl, "https://myserver/remote.php/dav/files")
	assert.Equal(t, nextcloudSource.FilePath, "Documents/Finances/RobTest.xlsx")
	assert.Equal(t, nextcloudSource.CostNameColumn, "D")
	assert.Equal(t, nextcloudSource.CostValueColumn, "E")
	assert.Equal(t, nextcloudSource.EarningNameColumn, "A")
	assert.Equal(t, nextcloudSource.EarningsValueColumn, "B")
	assert.Equal(t, nextcloudSource.StartRow, 2)
}
