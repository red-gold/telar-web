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

// FindNotificationsReceiver get all notifications by filter including receiver profile
func (s NotificationServiceImpl) FindNotificationsReceiver(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.Notification, error) {
	var pipeline []interface{}

	matchOperator := make(map[string]interface{})
	matchOperator["$match"] = filter

	sortOperator := make(map[string]interface{})
	sortOperator["$sort"] = sort

	pipeline = append(pipeline, matchOperator, sortOperator)

	if skip > 0 {
		skipOperator := make(map[string]interface{})
		skipOperator["$skip"] = skip
		pipeline = append(pipeline, skipOperator)
	}

	if limit > 0 {
		limitOperator := make(map[string]interface{})
		limitOperator["$limit"] = limit
		pipeline = append(pipeline, limitOperator)
	}

	lookupOperator := make(map[string]interface{})
	lookupOperator["$lookup"] = map[string]string{
		"localField":   "notifyRecieverUserId",
		"from":         "userProfile",
		"foreignField": "objectId",
		"as":           "userinfo",
	}

	unwindOperator := make(map[string]interface{})
	unwindOperator["$unwind"] = "$userinfo"

	projectOperator := make(map[string]interface{})
	project := make(map[string]interface{})

	project["objectId"] = 1
	project["ownerUserId"] = 1
	project["ownerDisplayName"] = 1
	project["ownerAvatar"] = 1
	project["created_date"] = 1
	project["description"] = 1
	project["url"] = 1
	project["notifyRecieverUserId"] = 1
	project["notifyRecieverEmail"] = "$userinfo.email"
	project["targetId"] = 1
	project["isSeen"] = 1
	project["type"] = 1
	project["emailNotification"] = 1
	project["isEmailSent"] = 1

	projectOperator["$project"] = project

	pipeline = append(pipeline, lookupOperator, unwindOperator, projectOperator)

	result := <-s.NotificationRepo.Aggregate(notificationCollectionName, pipeline)
	defer result.Close()
	if result.Error() != nil {
		return nil, result.Error()
	}
	var commentList []dto.Notification
	for result.Next() {
		var comment dto.Notification
		errDecode := result.Decode(&comment)
		if errDecode != nil {
			return nil, fmt.Errorf("Error docoding on dto.Comment")
		}
		commentList = append(commentList, comment)
	}

	return commentList, nil
}

// GetLastNotifications find by owner user id
func (s NotificationServiceImpl) GetLastNotifications() ([]dto.Notification, error) {
	sortMap := make(map[string]int)
	sortMap["created_date"] = -1
	filter := make(map[string]interface{})
	ne := make(map[string]interface{})
	ne["$ne"] = true
	filter["isEmailSent"] = ne
	return s.FindNotificationsReceiver(filter, 10, 0, sortMap)
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

// UpdateManyNotifications update many notifications
func (s NotificationServiceImpl) UpdateManyNotifications(filter interface{}, data interface{}, opts ...*coreData.UpdateOptions) error {

	result := <-s.NotificationRepo.UpdateMany(notificationCollectionName, filter, data, opts...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateBulkNotification update bulk notification
func (s NotificationServiceImpl) UpdateBulkNotification(bulk []data.BulkUpdateOne) error {

	result := <-s.NotificationRepo.BulkUpdateOne(notificationCollectionName, bulk)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateBulkNotificationList update bulk notification list
func (s NotificationServiceImpl) UpdateBulkNotificationList(userNotification []dto.Notification) error {
	var bulkList []repo.BulkUpdateOne
	for _, notification := range userNotification {
		filter := struct {
			ObjectId    uuid.UUID `json:"objectId" bson:"objectId"`
			OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
		}{
			ObjectId:    notification.ObjectId,
			OwnerUserId: notification.OwnerUserId,
		}

		setOperation := make(map[string]interface{})
		setOperation["$set"] = notification
		bulkItem := repo.BulkUpdateOne{
			Filter: filter,
			Data:   setOperation,
		}
		bulkList = append(bulkList, bulkItem)
	}
	return s.UpdateBulkNotification(bulkList)
}

// UpdateEmailSent update bulk notification list
func (s NotificationServiceImpl) UpdateEmailSent(notifyIds []uuid.UUID) error {

	include := make(map[string]interface{})
	include["$in"] = notifyIds

	filter := make(map[string]interface{})
	filter["objectId"] = include

	updateOperator := coreData.UpdateOperator{
		Set: map[string]bool{
			"isEmailSent": true,
		},
	}
	err := s.UpdateManyNotifications(filter, updateOperator)
	if err != nil {
		return err
	}
	return nil

}

// UpdateNotificationById update the notification
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
