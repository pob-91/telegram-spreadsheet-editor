package services

import (
	"encoding/json"
	"fmt"
	"io"
	e "nextcloud-spreadsheet-editor/errors"
	"nextcloud-spreadsheet-editor/model"
	"os"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type IMessagingService interface {
	GetCommandFromMessage(request io.ReadCloser) (*model.Command, error)
	SendTextMessage(chatId int64, message string) error
}

type TelegramService struct {
	Bot *tgbotapi.BotAPI
}

const (
	ALLOWED_USERS_KEY string = "TELEGRAM_ALLOWED_USERS"
)

func (s *TelegramService) GetCommandFromMessage(request io.ReadCloser) (*model.Command, error) {
	var update tgbotapi.Update
	if err := json.NewDecoder(request).Decode(&update); err != nil {
		zap.L().DPanic("Failed to parse telegram update", zap.Error(err))
		return nil, fmt.Errorf("Failed to parse telegram update")
	}

	allowedUsers := os.Getenv(ALLOWED_USERS_KEY)
	users := strings.Split(allowedUsers, ",")
	if !slices.Contains(users, fmt.Sprintf("%d", update.Message.From.ID)) {
		return nil, &e.CommandError{
			Unauthorized:    true,
			ChatId:          update.Message.Chat.ID,
			ResponseMessage: "",
		}
	}

	command, err := model.CommandFromMessage(update.Message.Text, update.Message.Chat.ID)
	if err != nil {
		return nil, err
	}

	return command, nil
}

func (s *TelegramService) SendTextMessage(chatId int64, message string) error {
	response := tgbotapi.NewMessage(chatId, message)
	if _, err := s.Bot.Send(response); err != nil {
		zap.L().Error("Failed to send telegram text message", zap.Error(err))
		return fmt.Errorf("Failed to sent bot message")
	}

	return nil
}
