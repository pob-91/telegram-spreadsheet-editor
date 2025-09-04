package model

import (
	"fmt"
	e "nextcloud-spreadsheet-editor/errors"
	"nextcloud-spreadsheet-editor/utils"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const (
	COMMAND_TYPE_LIST                   int = iota
	COMMAND_TYPE_UPDATE                 int = iota
	COMMAND_TYPE_UPDATE_CATEGORY_CHOSEN int = iota
	COMMAND_TYPE_NUMERICAL_AMOUNT       int = iota
	COMMAND_TYPE_UPDATE_FULL            int = iota
	COMMAND_TYPE_PING                   int = iota
)

type UpdateData struct {
	Category *string  `json:"category,omitempty"`
	Value    *float32 `json:"value,omitempty"`
}

type Command struct {
	Type      int   `json:"type"`
	UserId    int64 `json:"userId"`
	ChatId    int64 `json:"chatId"`
	MessageId int   `json:"messageId"`

	UpdateData *UpdateData `json:"updateData,omitempty"`
}

func CommandFromMessage(message string, chatId int64, messageId int, userId int64) (*Command, error) {
	norm := strings.ToLower(strings.ReplaceAll(message, " ", ""))
	switch {
	case norm == "ping":
		return &Command{
			Type:      COMMAND_TYPE_PING,
			ChatId:    chatId,
			MessageId: messageId,
			UserId:    userId,
		}, nil
	case norm == "list":
		return &Command{
			Type:      COMMAND_TYPE_LIST,
			ChatId:    chatId,
			MessageId: messageId,
			UserId:    userId,
		}, nil
	case norm == "update":
		return &Command{
			Type:      COMMAND_TYPE_UPDATE,
			ChatId:    chatId,
			MessageId: messageId,
			UserId:    userId,
		}, nil
	case utils.IsFinancial(norm):
		return commandFromFinancial(norm, chatId, messageId, userId)
	default:
		return nil, &e.CommandError{
			ResponseMessage: fmt.Sprintf("%s not a recognised command", message),
			ChatId:          chatId,
		}
	}
}

func CommandFromCallback(data string, chatId int64, messageId int, userId int64) (*Command, error) {
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
			UserId:    userId,
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

func MergeUpdateCommandWithFinancial(update *Command, financial *Command) *Command {
	return &Command{
		Type:      COMMAND_TYPE_UPDATE_FULL,
		ChatId:    update.ChatId,
		MessageId: update.MessageId,
		UserId:    update.UserId,
		UpdateData: &UpdateData{
			Category: update.UpdateData.Category,
			Value:    financial.UpdateData.Value,
		},
	}
}

// private

func commandFromFinancial(str string, chatId int64, messageId int, userId int64) (*Command, error) {
	// strip any £
	stripped := strings.ReplaceAll(str, "£", "")
	val, err := strconv.ParseFloat(stripped, 32)
	if err != nil {
		zap.L().DPanic("Failed to parse financial amount", zap.Error(err), zap.String("amount", stripped))
		return nil, &e.CommandError{
			ResponseMessage: fmt.Sprintf("Could not convert %s to GBP. Please enter a valid amount", str),
			ChatId:          chatId,
		}
	}

	lower := float32(val)
	return &Command{
		Type:      COMMAND_TYPE_NUMERICAL_AMOUNT,
		ChatId:    chatId,
		MessageId: messageId,
		UserId:    userId,
		UpdateData: &UpdateData{
			Value: &lower,
		},
	}, nil
}
