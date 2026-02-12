package services

import (
	"fmt"
	e "jarvis/tool_spreadsheet_editor/errors"
	"jarvis/tool_spreadsheet_editor/model"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type IMessagingService interface {
	GetCommandFromMessage(m *model.Message) (*model.Command, error)
	SendTextMessage(m *model.Message, chatId int64, message string) error
	SendEntryList(m *model.Message, chatId int64, entries *[]model.Entry) error
	SendCategorySelectionKeyboard(m *model.Message, chatId int64, entries *[]model.Entry, command string) error
	RemoveMarkupFromMessage(m *model.Message, chatId int64, messageId int) error
}

type TelegramService struct{}

func (s *TelegramService) GetCommandFromMessage(message *model.Message) (*model.Command, error) {
	if message.TelegramMessage == nil {
		zap.L().DPanic("Unhandled message source - telegram currently only input available")
		return nil, fmt.Errorf("Unhandled message source - telegram currently only input available")
	}

	update := message.TelegramMessage.Update
	bot := message.TelegramMessage.Bot

	var userId int64
	if update.CallbackQuery != nil {
		userId = update.CallbackQuery.From.ID
	} else {
		userId = update.Message.From.ID
	}

	if message.TelegramMessage.UserId != userId {
		zap.L().Warn("User not allowed", zap.Int64("userId", userId))
		return nil, &e.CommandError{
			Unauthorized:    true,
			ChatId:          update.Message.Chat.ID,
			ResponseMessage: "",
		}
	}

	if update.CallbackQuery != nil {
		// this is a response to something like an inline keyboard so process
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			zap.L().Error("Failed to respond to telegram callback", zap.Error(err))
			return nil, &e.CommandError{
				ChatId:          update.CallbackQuery.Message.Chat.ID,
				ResponseMessage: "Failed sorry...",
			}
		}

		command, err := model.CommandFromCallback(update.CallbackQuery.Data, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, userId)
		if err != nil {
			return nil, err
		}

		return command, nil
	}

	command, err := model.CommandFromMessage(update.Message.Text, update.Message.Chat.ID, update.Message.MessageID, userId)
	if err != nil {
		return nil, err
	}

	return command, nil
}

func (s *TelegramService) SendTextMessage(m *model.Message, chatId int64, message string) error {
	msg := tgbotapi.NewMessage(chatId, message)
	if _, err := m.TelegramMessage.Bot.Send(msg); err != nil {
		zap.L().Error("Failed to send telegram text message", zap.Error(err))
		return fmt.Errorf("Failed to sent bot message")
	}

	return nil
}

func (s *TelegramService) SendEntryList(m *model.Message, chatId int64, entries *[]model.Entry) error {
	var builder strings.Builder
	for _, e := range *entries {
		fmt.Fprintf(&builder, "%s %s\n", e.Category, e.Value)
	}

	msg := tgbotapi.NewMessage(chatId, builder.String())
	if _, err := m.TelegramMessage.Bot.Send(msg); err != nil {
		zap.L().Error("Failed to send telegram entries message", zap.Error(err))
		return fmt.Errorf("Failed to send bot entries message")
	}

	return nil
}

func (s *TelegramService) SendCategorySelectionKeyboard(m *model.Message, chatId int64, entries *[]model.Entry, command string) error {
	currentButtons := []tgbotapi.InlineKeyboardButton{}
	buttonRows := [][]tgbotapi.InlineKeyboardButton{}

	for i, e := range *entries {
		currentButtons = append(currentButtons, tgbotapi.NewInlineKeyboardButtonData(e.Category, fmt.Sprintf("%s:%s", command, e.Category)))
		if (i+1)%3 == 0 {
			buttonRows = append(buttonRows, currentButtons)
			currentButtons = make([]tgbotapi.InlineKeyboardButton, 0)
		}
	}

	msg := tgbotapi.NewMessage(chatId, "Please choose a category:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttonRows...)

	if _, err := m.TelegramMessage.Bot.Send(msg); err != nil {
		zap.L().Error("Failed to send telegram categories selection message", zap.Error(err))
		return fmt.Errorf("Failed to send bot categories selection message")
	}

	return nil
}

func (s *TelegramService) RemoveMarkupFromMessage(m *model.Message, chatId int64, messageId int) error {
	edit := tgbotapi.NewEditMessageReplyMarkup(chatId, messageId, tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{}))
	if _, err := m.TelegramMessage.Bot.Send(edit); err != nil {
		zap.L().Error("Failed to clear telegram markup", zap.Error(err))
		return fmt.Errorf("Failed to clear bot markup")
	}

	return nil
}
