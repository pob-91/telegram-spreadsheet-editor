package services

import "nextcloud-spreadsheet-editor/model"

type IStorageService interface{}

type RedisStorageService struct{}

func (s *RedisStorageService) StoreCommand(command *model.Command, userId int64) error {
	return nil
}

func (s *RedisStorageService) GetPreviousCommand(userId int64) (*model.Command, error) {
	return nil, nil
}
