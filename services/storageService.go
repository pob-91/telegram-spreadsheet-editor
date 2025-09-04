package services

import (
	"context"
	"encoding/json"
	"fmt"
	"nextcloud-spreadsheet-editor/errors"
	"nextcloud-spreadsheet-editor/model"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type IStorageService interface {
	StoreCommand(command *model.Command, userId int64) error
	GetPreviousCommand(userId int64) (*model.Command, error)
}

type RedisStorageService struct {
	Client *redis.Client
}

const (
	REDIS_HOST_KEY string = "REDIS_HOST"
)

func NewRedisStorageService() *RedisStorageService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv(REDIS_HOST_KEY),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &RedisStorageService{
		Client: rdb,
	}
}

func (s *RedisStorageService) StoreCommand(command *model.Command, userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	json, err := json.Marshal(command)
	if err != nil {
		zap.L().DPanic("Failed to serialise command", zap.Error(err))
		return fmt.Errorf("Failed to serialise command")
	}

	if err := s.Client.Set(ctx, strconv.FormatInt(userId, 10), string(json), time.Minute*15).Err(); err != nil {
		zap.L().Error("Failed to set command for user", zap.Error(err))
		return fmt.Errorf("Failed to set command for user")
	}

	return nil
}

func (s *RedisStorageService) GetPreviousCommand(userId int64) (*model.Command, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	jsonStr, err := s.Client.Get(ctx, strconv.FormatInt(userId, 10)).Result()
	if err != nil {
		if err == redis.Nil {
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
