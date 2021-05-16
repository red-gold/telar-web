package service

import (
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/data"
	dto "github.com/red-gold/telar-web/micros/setting/dto"
)

type UserSettingService interface {
	SaveUserSetting(userUserSetting *dto.UserSetting) error
	SaveManyUserSetting(userSettings []dto.UserSetting) error
	FindOneUserSetting(filter interface{}) (*dto.UserSetting, error)
	FindUserSettingList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserSetting, error)
	QueryUserSetting(search string, ownerUserId *uuid.UUID, userUserSettingTypeId *int, sortBy string, page int64) ([]dto.UserSetting, error)
	FindById(objectId uuid.UUID) (*dto.UserSetting, error)
	FindByOwnerUserId(ownerUserId uuid.UUID) ([]dto.UserSetting, error)
	FindSettingByUserIds(userIds []uuid.UUID, settingType string) ([]dto.UserSetting, error)
	UpdateBulkUserSetting(bulk []data.BulkUpdateOne) error
	UpdateUserSetting(filter interface{}, data interface{}) error
	UpdateUserSettingById(data *dto.UserSetting) error
	UpdateUserSettingsById(ownerUserId uuid.UUID, userSettings []dto.UserSetting) error
	DeleteUserSetting(filter interface{}) error
	DeleteUserSettingByOwner(ownerUserId uuid.UUID, userUserSettingId uuid.UUID) error
	DeleteManyUserSetting(filter interface{}) error
	CreateUserSettingIndex(indexes map[string]interface{}) error
	GetAllUserSetting(ownerUserId uuid.UUID) ([]dto.UserSetting, error)
	GetAllUserSettingByType(ownerUserId uuid.UUID, settingType string) ([]dto.UserSetting, error)
	DeleteUserSettingByOwnerUserId(ownerUserId uuid.UUID) error
}
