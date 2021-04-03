package service

import (
	"fmt"

	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/config"
	repo "github.com/red-gold/telar-core/data"
	"github.com/red-gold/telar-core/data/mongodb"
	mongoRepo "github.com/red-gold/telar-core/data/mongodb"
	"github.com/red-gold/telar-core/utils"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
)

// UserAuthService handlers with injected dependencies
type UserAuthServiceImpl struct {
	UserAuthRepo repo.Repository
}

// NewUserAuthService initializes UserAuthService's dependencies and create new UserAuthService struct
func NewUserAuthService(db interface{}) (UserAuthService, error) {

	userAuthService := &UserAuthServiceImpl{}

	switch *config.AppConfig.DBType {
	case config.DB_MONGO:

		mongodb := db.(mongodb.MongoDatabase)
		userAuthService.UserAuthRepo = mongoRepo.NewDataRepositoryMongo(mongodb)

	}
	if userAuthService.UserAuthRepo == nil {
		fmt.Printf("userAuthService.UserAuthRepo is nil! \n")
	}
	return userAuthService, nil
}

// SaveUserAuth save user authentication informaition
func (s UserAuthServiceImpl) SaveUserAuth(userAuth *dto.UserAuth) error {

	if userAuth.ObjectId == uuid.Nil {
		var uuidErr error
		userAuth.ObjectId, uuidErr = uuid.NewV4()
		if uuidErr != nil {
			return uuidErr
		}
	}

	if userAuth.CreatedDate == 0 {
		userAuth.CreatedDate = utils.UTCNowUnix()
	}

	result := <-s.UserAuthRepo.Save(userAuthCollectionName, userAuth)

	return result.Error
}

// FindOneUserAuth get all user authentication informaition
func (s UserAuthServiceImpl) FindOneUserAuth(filter interface{}) (*dto.UserAuth, error) {

	result := <-s.UserAuthRepo.FindOne(userAuthCollectionName, filter)
	if result.Error() != nil {
		if result.Error() == repo.ErrNoDocuments {
			return nil, nil
		}
		return nil, result.Error()
	}

	var userAuthResult dto.UserAuth
	errDecode := result.Decode(&userAuthResult)
	if errDecode != nil {
		return nil, fmt.Errorf("Error docoding on dto.UserAuth")
	}
	return &userAuthResult, nil
}

// FindUserAuthList get all user authentication informaition
func (s UserAuthServiceImpl) FindUserAuthList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserAuth, error) {

	result := <-s.UserAuthRepo.Find(userAuthCollectionName, filter, limit, skip, sort)
	defer result.Close()
	if result.Error() != nil {
		return nil, result.Error()
	}
	var userAuthList []dto.UserAuth
	for result.Next() {
		var userAuth dto.UserAuth
		errDecode := result.Decode(&userAuth)
		if errDecode != nil {
			return nil, fmt.Errorf("Error docoding on dto.UserAuth")
		}
		userAuthList = append(userAuthList, userAuth)
	}

	return userAuthList, nil
}

// FindByUsername find user auth by name
func (s UserAuthServiceImpl) FindByUsername(username string) (*dto.UserAuth, error) {

	filter := struct {
		Username string `json:"username" bson:"username"`
	}{
		Username: username,
	}
	return s.FindOneUserAuth(filter)
}

// FindByUserId find user auth by userId
func (s UserAuthServiceImpl) FindByUserId(userId uuid.UUID) (*dto.UserAuth, error) {

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: userId,
	}
	return s.FindOneUserAuth(filter)
}

// UpdateUserAuth update user auth information
func (s UserAuthServiceImpl) UpdateUserAuth(filter interface{}, data interface{}) error {

	result := <-s.UserAuthRepo.Update(userAuthCollectionName, filter, data)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdatePassword update user password
func (s UserAuthServiceImpl) UpdatePassword(userId uuid.UUID, newPassword []byte) error {

	updateData := struct {
		Set interface{} `json:"$set" bson:"$set"`
	}{
		Set: struct {
			Password []byte `json:"password" bson:"password"`
		}{
			Password: newPassword,
		},
	}

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: userId,
	}
	updateErr := s.UpdateUserAuth(filter, &updateData)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

// DeleteUserAuth get all user authentication informaition
func (s UserAuthServiceImpl) DeleteUserAuth(filter interface{}) error {

	result := <-s.UserAuthRepo.Delete(userAuthCollectionName, filter, true)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteManyUserAuth get all user authentication informaition
func (s UserAuthServiceImpl) DeleteManyUserAuth(filter interface{}) error {

	result := <-s.UserAuthRepo.Delete(userAuthCollectionName, filter, false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// CheckAdmin find user auth by userId
func (s UserAuthServiceImpl) CheckAdmin() (*dto.UserAuth, error) {

	filter := struct {
		Role string `json:"role" bson:"role"`
	}{
		Role: "admin",
	}
	return s.FindOneUserAuth(filter)
}
