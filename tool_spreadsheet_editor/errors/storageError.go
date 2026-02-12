package errors

const (
	STORAGE_ERROR_TYPE_NOT_FOUND int = iota
)

type StorageError struct {
	Type int
}

func (e *StorageError) Error() string {
	return "Storage error"
}
