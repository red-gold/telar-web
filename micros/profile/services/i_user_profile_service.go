package service

import (
	uuid "github.com/gofrs/uuid"
	dto "github.com/red-gold/telar-web/micros/profile/dto"
)

type UserProfileService interface {
	SaveUserProfile(userProfile *dto.UserProfile) error
	FindOneUserProfile(filter interface{}) (*dto.UserProfile, error)
	FindUserProfileList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserProfile, error)
	QueryUserProfile(search string, sortBy string, page int64, notIncludeUserIDList []uuid.UUID) ([]dto.UserProfile, error)
	FindProfileByUserIds(userIds []uuid.UUID) ([]dto.UserProfile, error)
	FindByUserId(userId uuid.UUID) (*dto.UserProfile, error)
	UpdateUserProfile(filter interface{}, data interface{}) error
	UpdateLastSeenNow(userId uuid.UUID) error
	UpdateUserProfileById(userId uuid.UUID, data *dto.UserProfile) error
	DeleteUserProfile(filter interface{}) error
	DeleteManyUserProfile(filter interface{}) error
	FindByUsername(username string) (*dto.UserProfile, error)
	CreateUserProfileIndex(indexes map[string]interface{}) error
	IncreaseFollowCount(objectId uuid.UUID, inc int) error
	IncreaseFollowerCount(objectId uuid.UUID, inc int) error
}
