package service

import (
	uuid "github.com/gofrs/uuid"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
)

type UserAuthService interface {
	SaveUserAuth(userAuth *dto.UserAuth) error
	FindOneUserAuth(filter interface{}) (*dto.UserAuth, error)
	FindUserAuthList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserAuth, error)
	FindByUserId(userId uuid.UUID) (*dto.UserAuth, error)
	UpdateUserAuth(filter interface{}, data interface{}) error
	UpdatePassword(userId uuid.UUID, newPassword []byte) error
	DeleteUserAuth(filter interface{}) error
	DeleteManyUserAuth(filter interface{}) error
	FindByUsername(username string) (*dto.UserAuth, error)
	CheckAdmin() (*dto.UserAuth, error)
}
