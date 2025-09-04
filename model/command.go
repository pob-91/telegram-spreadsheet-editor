package model

import (
	"fmt"
	e "nextcloud-spreadsheet-editor/errors"
	"strings"

	"go.uber.org/zap"
)

const (
	COMMAND_TYPE_LIST                   int = iota
	COMMAND_TYPE_UPDATE                 int = iota
	COMMAND_TYPE_UPDATE_CATEGORY_CHOSEN int = iota
)

type UpdateData struct {
	Category *string  `json:"category,omitempty"`
	Value    *float32 `json:"value,omitempty"`
}

type Command struct {
	Type      int   `json:"type"`
	ChatId    int64 `json:"chatId"`
	MessageId int   `json:"messageId"`

	UpdateData *UpdateData `json:"updateData,omitempty"`
}

func CommandFromMessage(message string, chatId int64, messageId int) (*Command, error) {
	switch strings.ToLower(strings.ReplaceAll(message, " ", "")) {
	case "list":
		return &Command{
			Type:      COMMAND_TYPE_LIST,
			ChatId:    chatId,
			MessageId: messageId,
		}, nil
	case "update":
		return &Command{
			Type:      COMMAND_TYPE_UPDATE,
			ChatId:    chatId,
			MessageId: messageId,
		}, nil
	default:
		return nil, &e.CommandError{
			ResponseMessage: fmt.Sprintf("%s not a recognised command", message),
			ChatId:          chatId,
		}
	}
}

func CommandFromCallback(data string, chatId int64, messageId int) (*Command, error) {
	// all callback data has the original command and then the data
	split := strings.Split(data, ":")
	if len(split) != 2 {
		zap.L().DPanic("Cannot parse callback data", zap.String("data", data))
		return nil, &e.CommandError{
			ResponseMessage: "Error, sorry...",
			ChatId:          chatId,
		}
	}

	category := split[1]

	switch split[0] {
	case "UPDATE":
		return &Command{
			Type:      COMMAND_TYPE_UPDATE_CATEGORY_CHOSEN,
			ChatId:    chatId,
			MessageId: messageId,
			UpdateData: &UpdateData{
				Category: &category,
			},
		}, nil
	default:
		return nil, &e.CommandError{
			ResponseMessage: fmt.Sprintf("%s not a recognised command", split[0]),
			ChatId:          chatId,
		}
	}
}
