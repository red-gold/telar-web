package service

import (
	"fmt"

	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/data"
	coreData "github.com/red-gold/telar-core/data"
	repo "github.com/red-gold/telar-core/data"
	"github.com/red-gold/telar-core/data/mongodb"
	mongoRepo "github.com/red-gold/telar-core/data/mongodb"
	"github.com/red-gold/telar-core/utils"
	dto "github.com/red-gold/telar-web/micros/setting/dto"
)

// UserSettingService handlers with injected dependencies
type UserSettingServiceImpl struct {
	UserSettingRepo repo.Repository
}

// NewUserSettingService initializes UserSettingService's dependencies and create new UserSettingService struct
func NewUserSettingService(db interface{}) (UserSettingService, error) {

	userSettingService := &UserSettingServiceImpl{}

	switch *config.AppConfig.DBType {
	case config.DB_MONGO:

		mongodb := db.(mongodb.MongoDatabase)
		userSettingService.UserSettingRepo = mongoRepo.NewDataRepositoryMongo(mongodb)

	}

	return userSettingService, nil
}

// SaveUserSetting save the userSetting
func (s UserSettingServiceImpl) SaveUserSetting(userSetting *dto.UserSetting) error {

	if userSetting.ObjectId == uuid.Nil {
		var uuidErr error
		userSetting.ObjectId, uuidErr = uuid.NewV4()
		if uuidErr != nil {
			return uuidErr
		}
	}

	if userSetting.CreatedDate == 0 {
		userSetting.CreatedDate = utils.UTCNowUnix()
	}

	result := <-s.UserSettingRepo.Save(userSettingCollectionName, userSetting)

	return result.Error
}

// SaveManyUserSetting save the userSetting
func (s UserSettingServiceImpl) SaveManyUserSetting(userSettings []dto.UserSetting) error {

	// https://github.com/golang/go/wiki/InterfaceSlice
	var interfaceSlice []interface{} = make([]interface{}, len(userSettings))
	for i, d := range userSettings {
		if d.ObjectId == uuid.Nil {
			var uuidErr error
			d.ObjectId, uuidErr = uuid.NewV4()
			if uuidErr != nil {
				return uuidErr
			}
		}

		if d.CreatedDate == 0 {
			d.CreatedDate = utils.UTCNowUnix()
		}
		interfaceSlice[i] = d
	}
	result := <-s.UserSettingRepo.SaveMany(userSettingCollectionName, interfaceSlice)

	return result.Error
}

// FindOneUserSetting get one userSetting
func (s UserSettingServiceImpl) FindOneUserSetting(filter interface{}) (*dto.UserSetting, error) {

	result := <-s.UserSettingRepo.FindOne(userSettingCollectionName, filter)
	if result.Error() != nil {
		return nil, result.Error()
	}

	var userSettingResult dto.UserSetting
	errDecode := result.Decode(&userSettingResult)
	if errDecode != nil {
		return nil, fmt.Errorf("Error docoding on dto.UserSetting")
	}
	return &userSettingResult, nil
}

// FindUserSettingList get all userSettings by filter
func (s UserSettingServiceImpl) FindUserSettingList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserSetting, error) {

	result := <-s.UserSettingRepo.Find(userSettingCollectionName, filter, limit, skip, sort)
	defer result.Close()
	if result.Error() != nil {
		return nil, result.Error()
	}
	var userSettingList []dto.UserSetting
	for result.Next() {
		var userSetting dto.UserSetting
		errDecode := result.Decode(&userSetting)
		if errDecode != nil {
			return nil, fmt.Errorf("Error docoding on dto.UserSetting")
		}
		userSettingList = append(userSettingList, userSetting)
	}

	return userSettingList, nil
}

// QueryUserSetting get all userSettings by query
func (s UserSettingServiceImpl) QueryUserSetting(search string, ownerUserId *uuid.UUID, userSettingTypeId *int, sortBy string, page int64) ([]dto.UserSetting, error) {
	sortMap := make(map[string]int)
	sortMap[sortBy] = -1
	skip := numberOfItems * (page - 1)
	limit := numberOfItems

	filter := make(map[string]interface{})
	if search != "" {
		filter["$text"] = coreData.SearchOperator{Search: search}
	}
	if ownerUserId != nil {
		filter["ownerUserId"] = *ownerUserId
	}
	if userSettingTypeId != nil {
		filter["userSettingTypeId"] = *userSettingTypeId
	}
	fmt.Println(filter)
	result, err := s.FindUserSettingList(filter, limit, skip, sortMap)

	return result, err
}

// FindByOwnerUserId find by owner user id
func (s UserSettingServiceImpl) FindByOwnerUserId(ownerUserId uuid.UUID) ([]dto.UserSetting, error) {
	sortMap := make(map[string]int)
	sortMap["created_date"] = -1
	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		OwnerUserId: ownerUserId,
	}
	return s.FindUserSettingList(filter, 0, 0, sortMap)
}

// FindById find by userSetting id
func (s UserSettingServiceImpl) FindById(objectId uuid.UUID) (*dto.UserSetting, error) {

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: objectId,
	}
	return s.FindOneUserSetting(filter)
}

// FindSettingByUserIds Find setting by user IDs
func (s UserSettingServiceImpl) FindSettingByUserIds(userIds []uuid.UUID, settingType string) ([]dto.UserSetting, error) {
	filter := make(map[string]interface{})
	sortMap := make(map[string]int)
	sortMap["createdDate"] = -1

	include := make(map[string]interface{})
	include["$in"] = userIds
	filter["ownerUserId"] = include

	if settingType != "" {
		equal := make(map[string]interface{})
		equal["$eq"] = settingType
		filter["type"] = equal
	}

	result, err := s.FindUserSettingList(filter, 0, 0, sortMap)

	return result, err
}

// UpdateUserSetting update the userSetting
func (s UserSettingServiceImpl) UpdateUserSetting(filter interface{}, data interface{}) error {

	result := <-s.UserSettingRepo.Update(userSettingCollectionName, filter, data)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateUserSetting update the userSetting
func (s UserSettingServiceImpl) UpdateBulkUserSetting(bulk []data.BulkUpdateOne) error {

	result := <-s.UserSettingRepo.BulkUpdateOne(userSettingCollectionName, bulk)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateUserSetting update the userSetting
func (s UserSettingServiceImpl) UpdateUserSettingsById(ownerUserId uuid.UUID, userSettings []dto.UserSetting) error {
	var bulkList []repo.BulkUpdateOne
	for _, setting := range userSettings {
		filter := struct {
			ObjectId    uuid.UUID `json:"objectId" bson:"objectId"`
			OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
		}{
			ObjectId:    setting.ObjectId,
			OwnerUserId: setting.OwnerUserId,
		}

		setOperation := make(map[string]interface{})
		setOperation["$set"] = setting
		bulkItem := repo.BulkUpdateOne{
			Filter: filter,
			Data:   setOperation,
		}
		bulkList = append(bulkList, bulkItem)
	}
	return s.UpdateBulkUserSetting(bulkList)
}

// UpdateUserSetting update the userSetting by objectId and ownerUserId
func (s UserSettingServiceImpl) UpdateUserSettingById(data *dto.UserSetting) error {
	filter := struct {
		ObjectId    uuid.UUID `json:"objectId" bson:"objectId"`
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		ObjectId:    data.ObjectId,
		OwnerUserId: data.OwnerUserId,
	}

	updateOperator := coreData.UpdateOperator{
		Set: data,
	}
	err := s.UpdateUserSetting(filter, updateOperator)
	if err != nil {
		return err
	}
	return nil
}

// DeleteUserSetting delete userSetting by filter
func (s UserSettingServiceImpl) DeleteUserSetting(filter interface{}) error {

	result := <-s.UserSettingRepo.Delete(userSettingCollectionName, filter, true)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteUserSetting delete userSetting by ownerUserId and userSettingId
func (s UserSettingServiceImpl) DeleteUserSettingByOwner(ownerUserId uuid.UUID, userSettingId uuid.UUID) error {

	filter := struct {
		ObjectId    uuid.UUID `json:"objectId" bson:"objectId"`
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		ObjectId:    userSettingId,
		OwnerUserId: ownerUserId,
	}
	err := s.DeleteUserSetting(filter)
	if err != nil {
		return err
	}
	return nil
}

// DeleteManyUserSetting delete many userSetting by filter
func (s UserSettingServiceImpl) DeleteManyUserSetting(filter interface{}) error {

	result := <-s.UserSettingRepo.Delete(userSettingCollectionName, filter, false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// CreateUserSettingIndex create index for userSetting search.
func (s UserSettingServiceImpl) CreateUserSettingIndex(indexes map[string]interface{}) error {
	result := <-s.UserSettingRepo.CreateIndex(userSettingCollectionName, indexes)
	return result
}

// GetAllUserSetting get all user setting
func (s UserSettingServiceImpl) GetAllUserSetting(ownerUserId uuid.UUID) ([]dto.UserSetting, error) {
	sortMap := make(map[string]int)
	sortMap["created_date"] = -1
	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		OwnerUserId: ownerUserId,
	}
	return s.FindUserSettingList(filter, 0, 0, sortMap)
}

// GetUserSettingByType get all user setting by setting type
func (s UserSettingServiceImpl) GetAllUserSettingByType(ownerUserId uuid.UUID, settingType string) ([]dto.UserSetting, error) {
	sortMap := make(map[string]int)
	sortMap["created_date"] = -1
	filter := struct {
		Type        string    `json:"type" bson:"type"`
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		Type:        settingType,
		OwnerUserId: ownerUserId,
	}
	return s.FindUserSettingList(filter, 0, 0, sortMap)
}

// DeleteUserSetting delete userSetting by ownerUserId
func (s UserSettingServiceImpl) DeleteUserSettingByOwnerUserId(ownerUserId uuid.UUID) error {

	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		OwnerUserId: ownerUserId,
	}
	err := s.DeleteManyUserSetting(filter)
	if err != nil {
		return err
	}
	return nil
}
