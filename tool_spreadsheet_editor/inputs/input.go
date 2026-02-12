package inputs

import "jarvis/tool_spreadsheet_editor/model"

type Input interface {
	GetType() string
	Start(handler func(*model.Message))
	Stop()
}
