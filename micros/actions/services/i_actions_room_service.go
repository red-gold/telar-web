package service

import (
	uuid "github.com/gofrs/uuid"
	coreData "github.com/red-gold/telar-core/data"
	dto "github.com/red-gold/telar-web/micros/actions/dto"
)

type ActionRoomService interface {
	SaveActionRoom(actionRoom *dto.ActionRoom) error
	FindOneActionRoom(filter interface{}) (*dto.ActionRoom, error)
	FindActionRoomList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.ActionRoom, error)
	FindById(objectId uuid.UUID) (*dto.ActionRoom, error)
	FindByOwnerUserId(ownerUserId uuid.UUID) ([]dto.ActionRoom, error)
	UpdateActionRoom(filter interface{}, data interface{}, opts ...*coreData.UpdateOptions) error
	UpdateActionRoomById(data *dto.ActionRoom) error
	DeleteActionRoom(filter interface{}) error
	DeleteActionRoomByOwner(ownerUserId uuid.UUID, actionRoomId uuid.UUID) error
	DeleteManyActionRooms(filter interface{}) error
	CreateActionRoomIndex(indexes map[string]interface{}) error
	SetAccessKey(ownerUserId uuid.UUID) (string, error)
	VerifyAccessKey(ownerUserId uuid.UUID, accessKey string) (bool, error)
	GetAccessKey(ownerUserId uuid.UUID) (string, error)
}
