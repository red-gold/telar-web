package service

import (
	"fmt"

	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/config"
	coreData "github.com/red-gold/telar-core/data"
	repo "github.com/red-gold/telar-core/data"
	"github.com/red-gold/telar-core/data/mongodb"
	mongoRepo "github.com/red-gold/telar-core/data/mongodb"
	"github.com/red-gold/telar-core/utils"
	dto "github.com/red-gold/telar-web/micros/actions/dto"
)

// ActionRoomService handlers with injected dependencies
type ActionRoomServiceImpl struct {
	ActionRoomRepo repo.Repository
}

// NewActionRoomService initializes ActionRoomService's dependencies and create new ActionRoomService struct
func NewActionRoomService(db interface{}) (ActionRoomService, error) {

	actionRoomService := &ActionRoomServiceImpl{}

	switch *config.AppConfig.DBType {
	case config.DB_MONGO:

		mongodb := db.(mongodb.MongoDatabase)
		actionRoomService.ActionRoomRepo = mongoRepo.NewDataRepositoryMongo(mongodb)

	}

	return actionRoomService, nil
}

// SaveActionRoom save the actionRoom
func (s ActionRoomServiceImpl) SaveActionRoom(actionRoom *dto.ActionRoom) error {

	if actionRoom.ObjectId == uuid.Nil {
		var uuidErr error
		actionRoom.ObjectId, uuidErr = uuid.NewV4()
		if uuidErr != nil {
			return uuidErr
		}
	}

	if actionRoom.CreatedDate == 0 {
		actionRoom.CreatedDate = utils.UTCNowUnix()
	}

	result := <-s.ActionRoomRepo.Save(actionRoomCollectionName, actionRoom)

	return result.Error
}

// FindOneActionRoom get one actionRoom
func (s ActionRoomServiceImpl) FindOneActionRoom(filter interface{}) (*dto.ActionRoom, error) {

	result := <-s.ActionRoomRepo.FindOne(actionRoomCollectionName, filter)
	if result.Error() != nil {
		return nil, result.Error()
	}

	var actionRoomResult dto.ActionRoom
	errDecode := result.Decode(&actionRoomResult)
	if errDecode != nil {
		return nil, fmt.Errorf("Error docoding on dto.ActionRoom")
	}
	return &actionRoomResult, nil
}

// FindActionRoomList get all actionRooms by filter
func (s ActionRoomServiceImpl) FindActionRoomList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.ActionRoom, error) {

	result := <-s.ActionRoomRepo.Find(actionRoomCollectionName, filter, limit, skip, sort)
	defer result.Close()
	if result.Error() != nil {
		return nil, result.Error()
	}
	var actionRoomList []dto.ActionRoom
	for result.Next() {
		var actionRoom dto.ActionRoom
		errDecode := result.Decode(&actionRoom)
		if errDecode != nil {
			return nil, fmt.Errorf("Error docoding on dto.ActionRoom")
		}
		actionRoomList = append(actionRoomList, actionRoom)
	}

	return actionRoomList, nil
}

// FindByOwnerUserId find by owner user id
func (s ActionRoomServiceImpl) FindByOwnerUserId(ownerUserId uuid.UUID) ([]dto.ActionRoom, error) {
	sortMap := make(map[string]int)
	sortMap["created_date"] = -1
	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		OwnerUserId: ownerUserId,
	}
	return s.FindActionRoomList(filter, 0, 0, sortMap)
}

// FindById find by actionRoom id
func (s ActionRoomServiceImpl) FindById(objectId uuid.UUID) (*dto.ActionRoom, error) {

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: objectId,
	}
	return s.FindOneActionRoom(filter)
}

// UpdateActionRoom update the actionRoom
func (s ActionRoomServiceImpl) UpdateActionRoom(filter interface{}, data interface{}, opts ...*coreData.UpdateOptions) error {

	result := <-s.ActionRoomRepo.Update(actionRoomCollectionName, filter, data, opts...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateActionRoom update the actionRoom
func (s ActionRoomServiceImpl) UpdateActionRoomById(data *dto.ActionRoom) error {
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
	err := s.UpdateActionRoom(filter, updateOperator)
	if err != nil {
		return err
	}
	return nil
}

// DeleteActionRoom delete actionRoom by filter
func (s ActionRoomServiceImpl) DeleteActionRoom(filter interface{}) error {

	result := <-s.ActionRoomRepo.Delete(actionRoomCollectionName, filter, true)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteActionRoom delete actionRoom by ownerUserId and actionRoomId
func (s ActionRoomServiceImpl) DeleteActionRoomByOwner(ownerUserId uuid.UUID, actionRoomId uuid.UUID) error {

	filter := struct {
		ObjectId    uuid.UUID `json:"objectId" bson:"objectId"`
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		ObjectId:    actionRoomId,
		OwnerUserId: ownerUserId,
	}
	err := s.DeleteActionRoom(filter)
	if err != nil {
		return err
	}
	return nil
}

// DeleteManyActionRooms delete many actionRooms by filter
func (s ActionRoomServiceImpl) DeleteManyActionRooms(filter interface{}) error {

	result := <-s.ActionRoomRepo.Delete(actionRoomCollectionName, filter, false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// CreateActionRoomIndex create index for actionRoom search.
func (s ActionRoomServiceImpl) CreateActionRoomIndex(indexes map[string]interface{}) error {
	result := <-s.ActionRoomRepo.CreateIndex(actionRoomCollectionName, indexes)
	return result
}

// SetAccessKey create access key for action room
func (s ActionRoomServiceImpl) SetAccessKey(ownerUserId uuid.UUID) (string, error) {

	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		OwnerUserId: ownerUserId,
	}
	accessKey, uuidErr := uuid.NewV4()
	if uuidErr != nil {
		return "", uuidErr
	}
	updateOperator := coreData.UpdateOperator{
		Set: struct {
			AccessKey string `json:"accessKey" bson:"accessKey"`
		}{
			AccessKey: accessKey.String(),
		},
	}
	options := &coreData.UpdateOptions{}
	options.SetUpsert(true)
	updateErr := s.UpdateActionRoom(filter, updateOperator, options)
	if updateErr != nil {
		return "", updateErr
	}
	return accessKey.String(), nil
}

// VerifyAccessKey increment score of post
func (s ActionRoomServiceImpl) VerifyAccessKey(ownerUserId uuid.UUID, accessKey string) (bool, error) {

	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
		AccessKey   string    `json:"accessKey" bson:"accessKey"`
	}{
		AccessKey:   accessKey,
		OwnerUserId: ownerUserId,
	}

	foundActionRoom, findErr := s.FindOneActionRoom(filter)
	if findErr != nil {
		return false, findErr
	}

	return (foundActionRoom.ObjectId != uuid.Nil), nil
}

// GetAccessKey increment score of post
func (s ActionRoomServiceImpl) GetAccessKey(ownerUserId uuid.UUID) (string, error) {

	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		OwnerUserId: ownerUserId,
	}

	options := &coreData.UpdateOptions{}
	options.SetUpsert(true)
	foundActionRoom, findErr := s.FindOneActionRoom(filter)
	if findErr != nil {
		return "", findErr
	}

	return foundActionRoom.AccessKey, nil
}
