package model

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// This is just a simple wrapper around messages so we can use multiple input providers in the future.
type Message struct {
	TelegramMessage *tgbotapi.Update
}
