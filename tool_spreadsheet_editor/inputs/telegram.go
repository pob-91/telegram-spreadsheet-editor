package inputs

import (
	"fmt"
	"jarvis/tool_spreadsheet_editor/model"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type TelegramInput struct {
	Bot      *tgbotapi.BotAPI
	UserId   int64
	UserName string
}

func NewTelegramInput(input *model.TelegramInput, user string) (*TelegramInput, error) {
	token, exists := os.LookupEnv(input.TokenEnv)
	if !exists {
		zap.L().Panic("No telegram input token")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		zap.L().Error("Failed to init new bot API", zap.Error(err))
		return nil, fmt.Errorf("Failed to init new bot API: %w", err)
	}

	return &TelegramInput{
		Bot:      bot,
		UserId:   input.UserId,
		UserName: user,
	}, nil
}

// public

func (i *TelegramInput) GetType() string {
	return model.INPUT_TYPE_TELEGRAM
}

func (i *TelegramInput) Start(handler func(*model.Message)) {
	// check for webhooks
	if err := i.checkForWebhooks(); err != nil {
		return
	}

	// process signal messages
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := i.Bot.GetUpdatesChan(u)

	zap.L().Info("Starting telegram input", zap.String("user", i.UserName))

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}
		message := model.Message{
			UserName: i.UserName,
			TelegramMessage: &model.TelegramMessage{
				Update: &update,
				Bot:    i.Bot,
				UserId: i.UserId,
			},
		}
		handler(&message)
	}
}

func (i *TelegramInput) Stop() {
	i.Bot.StopReceivingUpdates()
}

// private

func (i *TelegramInput) checkForWebhooks() error {
	info, err := i.Bot.GetWebhookInfo()
	if err != nil {
		zap.L().Warn("Failed to check for existing webhooks.", zap.Error(err), zap.Int64("userId", i.UserId))
		return nil
	}

	if !info.IsSet() {
		return nil
	}

	cfg := tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: true,
	}
	if _, err := i.Bot.Request(cfg); err != nil {
		zap.L().DPanic("Failed to delete webhook - cannot proceed.", zap.Error(err), zap.Int64("userId", i.UserId))
		return fmt.Errorf("Failed to delete webhook - cannot proceed: %w", err)
	}

	return nil
}
