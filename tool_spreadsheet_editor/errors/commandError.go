package errors

type CommandError struct {
	Unauthorized    bool
	ResponseMessage string
	ChatId          int64
}

func (e *CommandError) Error() string {
	return e.ResponseMessage
}
