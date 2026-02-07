package inputs

import "telegram-spreadsheet-editor/model"

type Input interface {
	GetType() string
	Start(handler func(*model.Message))
	Stop()
}
