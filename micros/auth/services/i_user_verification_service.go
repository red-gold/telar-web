package service

import (
	uuid "github.com/gofrs/uuid"
	tsconfig "github.com/red-gold/telar-core/config"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
)

type UserVerificationService interface {
	SaveUserVerification(userAuth *dto.UserVerification) error
	FindOneUserVerification(filter interface{}) (*dto.UserVerification, error)
	FindUserVerificationList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserVerification, error)
	FindByUserId(userId uuid.UUID) (*dto.UserVerification, error)
	FindByVerifyId(verifyId uuid.UUID) (*dto.UserVerification, error)
	UpdateUserVerification(filter interface{}, data interface{}) error
	DeleteUserVerification(filter interface{}) error
	DeleteManyUserVerification(filter interface{}) error
	VerifyUserByCode(userId uuid.UUID, verifyId uuid.UUID, remoteIpAddress string, code string, target string) (bool, error)
	CreateEmailVerficationToken(input EmailVerificationToken,
		coreConfig *tsconfig.Configuration) (string, error)
	CreatePhoneVerficationToken(input PhoneVerificationToken,
		coreConfig *tsconfig.Configuration) (string, error)
}
