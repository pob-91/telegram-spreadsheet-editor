package services

import (
	"context"
	"encoding/json"
	"fmt"
	"jarvis/tool_spreadsheet_editor/errors"
	"jarvis/tool_spreadsheet_editor/model"
	"os"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
	"go.uber.org/zap"
)

type IStorageService interface {
	StoreCommand(command *model.Command, userId int64) error
	GetPreviousCommand(userId int64) (*model.Command, error)
}

type ValkeyStorageService struct {
	Client valkey.Client
}

const (
	VALKEY_HOST_KEY string = "VALKEY_HOST"
)

func NewValkeyStorageService() *ValkeyStorageService {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{
			os.Getenv(VALKEY_HOST_KEY),
		},
	})
	if err != nil {
		zap.L().Panic("Failed to init valkey client", zap.Error(err))
	}

	return &ValkeyStorageService{
		Client: client,
	}
}

func (s *ValkeyStorageService) StoreCommand(command *model.Command, userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	json, err := json.Marshal(command)
	if err != nil {
		zap.L().DPanic("Failed to serialise command", zap.Error(err))
		return fmt.Errorf("Failed to serialise command")
	}

	key := strconv.FormatInt(userId, 10)
	if err := s.Client.Do(ctx, s.Client.B().Set().Key(key).Value(string(json)).Build()).Error(); err != nil {
		zap.L().Error("Failed to set command for user", zap.Error(err))
		return fmt.Errorf("Failed to set command for user")
	}

	return nil
}

func (s *ValkeyStorageService) GetPreviousCommand(userId int64) (*model.Command, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	jsonStr, err := s.Client.Do(ctx, s.Client.B().Get().Key(strconv.FormatInt(userId, 10)).Build()).ToString()
	if err != nil {
		if err == valkey.Nil {
			return nil, &errors.StorageError{
				Type: errors.STORAGE_ERROR_TYPE_NOT_FOUND,
			}
		}
		zap.L().Error("Failed to get command for user", zap.Error(err))
		return nil, fmt.Errorf("Failed to get command for user")
	}

	// don't actually need this as we are always overwriting anyway but leaving here as a good exampe
	// if _, err := s.Client.Expire(ctx, strconv.FormatInt(userId, 10), time.Minute*15).Result(); err != nil {
	// 	zap.L().Error("Failed to update expiry for command", zap.Error(err))
	// 	return nil, fmt.Errorf("Failed to update expiry for command")
	// }

	var command model.Command
	if err := json.Unmarshal([]byte(jsonStr), &command); err != nil {
		zap.L().DPanic("Failed to unmarshal json command", zap.Error(err))
		return nil, fmt.Errorf("Failed to unmarshal json command")
	}

	return &command, nil
}
