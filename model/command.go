package model

import (
	"fmt"
	e "nextcloud-spreadsheet-editor/errors"
	"strings"
)

const (
	COMMAND_TYPE_LIST int = iota
)

type Command struct {
	Type   int
	ChatId int64
}

func CommandFromMessage(message string, chatId int64) (*Command, error) {
	switch strings.ToLower(strings.ReplaceAll(message, " ", "")) {
	case "list":
		return &Command{
			Type:   COMMAND_TYPE_LIST,
			ChatId: chatId,
		}, nil
	default:
		return nil, &e.CommandError{
			ResponseMessage: fmt.Sprintf("%s not a recognised command", message),
			ChatId:          chatId,
		}
	}
}
