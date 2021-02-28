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
	dto "github.com/red-gold/telar-web/micros/notifications/dto"
)

// NotificationService handlers with injected dependencies
type NotificationServiceImpl struct {
	NotificationRepo repo.Repository
}

// NewNotificationService initializes NotificationService's dependencies and create new NotificationService struct
func NewNotificationService(db interface{}) (NotificationService, error) {

	notificationService := &NotificationServiceImpl{}

	switch *config.AppConfig.DBType {
	case config.DB_MONGO:

		mongodb := db.(mongodb.MongoDatabase)
		notificationService.NotificationRepo = mongoRepo.NewDataRepositoryMongo(mongodb)

	}

	return notificationService, nil
}

// SaveNotification save the notification
func (s NotificationServiceImpl) SaveNotification(notification *dto.Notification) error {

	if notification.ObjectId == uuid.Nil {
		var uuidErr error
		notification.ObjectId, uuidErr = uuid.NewV4()
		if uuidErr != nil {
			return uuidErr
		}
	}

	if notification.CreatedDate == 0 {
		notification.CreatedDate = utils.UTCNowUnix()
	}

	result := <-s.NotificationRepo.Save(notificationCollectionName, notification)

	return result.Error
}

// FindOneNotification get one notification
func (s NotificationServiceImpl) FindOneNotification(filter interface{}) (*dto.Notification, error) {

	result := <-s.NotificationRepo.FindOne(notificationCollectionName, filter)
	if result.Error() != nil {
		return nil, result.Error()
	}

	var notificationResult dto.Notification
	errDecode := result.Decode(&notificationResult)
	if errDecode != nil {
		return nil, fmt.Errorf("Error docoding on dto.Notification")
	}
	return &notificationResult, nil
}

// FindNotificationList get all notifications by filter
func (s NotificationServiceImpl) FindNotificationList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.Notification, error) {

	result := <-s.NotificationRepo.Find(notificationCollectionName, filter, limit, skip, sort)
	defer result.Close()
	if result.Error() != nil {
		return nil, result.Error()
	}
	var notificationList []dto.Notification
	for result.Next() {
		var notification dto.Notification
		errDecode := result.Decode(&notification)
		if errDecode != nil {
			return nil, fmt.Errorf("Error docoding on dto.Notification")
		}
		notificationList = append(notificationList, notification)
	}

	return notificationList, nil
}

// FindByOwnerUserId find by owner user id
func (s NotificationServiceImpl) FindByOwnerUserId(ownerUserId uuid.UUID) ([]dto.Notification, error) {
	sortMap := make(map[string]int)
	sortMap["created_date"] = -1
	filter := struct {
		OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	}{
		OwnerUserId: ownerUserId,
	}
	return s.FindNotificationList(filter, 0, 0, sortMap)
}

// FindById find by notification id
func (s NotificationServiceImpl) FindById(objectId uuid.UUID) (*dto.Notification, error) {

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: objectId,
	}
	return s.FindOneNotification(filter)
}

// UpdateNotification update the notification
func (s NotificationServiceImpl) UpdateNotification(filter interface{}, data interface{}, opts ...*coreData.UpdateOptions) error {

	result := <-s.NotificationRepo.Update(notificationCollectionName, filter, data, opts...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateNotification update the notification
func (s NotificationServiceImpl) UpdateNotificationById(data *dto.Notification) error {
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
	err := s.UpdateNotification(filter, updateOperator)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNotification delete notification by filter
func (s NotificationServiceImpl) DeleteNotification(filter interface{}) error {

	result := <-s.NotificationRepo.Delete(notificationCollectionName, filter, true)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteNotification delete notification by notificationReceiverId and notificationId
func (s NotificationServiceImpl) DeleteNotificationByOwner(notificationReceiverId uuid.UUID, notificationId uuid.UUID) error {

	filter := struct {
		ObjectId             uuid.UUID `json:"objectId" bson:"objectId"`
		NotifyRecieverUserId uuid.UUID `json:"notifyRecieverUserId" bson:"notifyRecieverUserId"`
	}{
		ObjectId:             notificationId,
		NotifyRecieverUserId: notificationReceiverId,
	}
	err := s.DeleteNotification(filter)
	if err != nil {
		return err
	}
	return nil
}

// DeleteManyNotifications delete many notifications by filter
func (s NotificationServiceImpl) DeleteManyNotifications(filter interface{}) error {

	result := <-s.NotificationRepo.Delete(notificationCollectionName, filter, false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// CreateNotificationIndex create index for notification search.
func (s NotificationServiceImpl) CreateNotificationIndex(indexes map[string]interface{}) error {
	result := <-s.NotificationRepo.CreateIndex(notificationCollectionName, indexes)
	return result
}

// GetNotificationByUserId get all notifications by userId who receive the notification
func (s NotificationServiceImpl) GetNotificationByUserId(userId *uuid.UUID, sortBy string, page int64) ([]dto.Notification, error) {
	sortMap := make(map[string]int)
	sortMap[sortBy] = -1
	skip := numberOfItems * (page - 1)
	limit := numberOfItems

	filter := struct {
		NotifyRecieverUserId uuid.UUID `json:"notifyRecieverUserId" bson:"notifyRecieverUserId"`
	}{
		NotifyRecieverUserId: *userId,
	}
	result, err := s.FindNotificationList(filter, limit, skip, sortMap)

	return result, err
}

// SeenNotification update the notification to seen
func (s NotificationServiceImpl) SeenNotification(objectId uuid.UUID, userId uuid.UUID) error {
	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
		UserId   uuid.UUID `json:"notifyRecieverUserId" bson:"notifyRecieverUserId"`
	}{
		ObjectId: objectId,
		UserId:   userId,
	}

	updateOperator := coreData.UpdateOperator{
		Set: struct {
			IsSeen bool `json:"isSeen" bson:"isSeen"`
		}{
			IsSeen: true,
		},
	}
	err := s.UpdateNotification(filter, updateOperator)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNotificationsByUserId delete notifications by userId
func (s NotificationServiceImpl) DeleteNotificationsByUserId(userId uuid.UUID) error {

	filter := struct {
		UserId uuid.UUID `json:"notifyRecieverUserId" bson:"notifyRecieverUserId"`
	}{
		UserId: userId,
	}
	err := s.DeleteManyNotifications(filter)
	if err != nil {
		return err
	}
	return nil
}
