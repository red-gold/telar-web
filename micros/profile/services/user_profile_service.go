package service

import (
	"fmt"

	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/config"
	coreData "github.com/red-gold/telar-core/data"
	"github.com/red-gold/telar-core/data/mongodb"
	mongoRepo "github.com/red-gold/telar-core/data/mongodb"
	"github.com/red-gold/telar-core/utils"
	dto "github.com/red-gold/telar-web/micros/profile/dto"
)

// UserProfileService handlers with injected dependencies
type UserProfileServiceImpl struct {
	UserProfileRepo coreData.Repository
}

// NewUserProfileService initializes UserProfileService's dependencies and create new UserProfileService struct
func NewUserProfileService(db interface{}) (UserProfileService, error) {

	userProfileService := &UserProfileServiceImpl{}

	switch *config.AppConfig.DBType {
	case config.DB_MONGO:

		mongodb := db.(mongodb.MongoDatabase)
		userProfileService.UserProfileRepo = mongoRepo.NewDataRepositoryMongo(mongodb)

	}
	if userProfileService.UserProfileRepo == nil {
		fmt.Printf("userProfileService.UserProfileRepo is nil! \n")
	}
	return userProfileService, nil
}

// SaveUserProfile save user profile informaition
func (s UserProfileServiceImpl) SaveUserProfile(userProfile *dto.UserProfile) error {

	if userProfile.ObjectId == uuid.Nil {
		var uuidErr error
		userProfile.ObjectId, uuidErr = uuid.NewV4()
		if uuidErr != nil {
			return uuidErr
		}
	}

	if userProfile.CreatedDate == 0 {
		userProfile.CreatedDate = utils.UTCNowUnix()
	}

	result := <-s.UserProfileRepo.Save(userProfileCollectionName, userProfile)

	return result.Error
}

// FindOneUserProfile get one user profile informaition
func (s UserProfileServiceImpl) FindOneUserProfile(filter interface{}) (*dto.UserProfile, error) {

	result := <-s.UserProfileRepo.FindOne(userProfileCollectionName, filter)
	if result.Error() != nil {
		if result.Error() == coreData.ErrNoDocuments {
			return nil, nil
		}
		return nil, result.Error()
	}

	var userProfileResult dto.UserProfile
	errDecode := result.Decode(&userProfileResult)
	if errDecode != nil {
		return nil, fmt.Errorf("Error docoding on dto.UserProfile")
	}
	return &userProfileResult, nil
}

// FindUserProfileList get all user profile informaition
func (s UserProfileServiceImpl) FindUserProfileList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserProfile, error) {

	result := <-s.UserProfileRepo.Find("userProfile", filter, limit, skip, sort)
	defer result.Close()
	if result.Error() != nil {
		return nil, result.Error()
	}
	var userProfileList []dto.UserProfile
	for result.Next() {
		var userProfile dto.UserProfile
		errDecode := result.Decode(&userProfile)
		if errDecode != nil {
			return nil, fmt.Errorf("Error docoding on dto.UserProfile")
		}
		userProfileList = append(userProfileList, userProfile)
	}

	return userProfileList, nil
}

// QueryPost get all user profile by query
func (s UserProfileServiceImpl) QueryUserProfile(search string, sortBy string, page int64, notIncludeUserIDList []uuid.UUID) ([]dto.UserProfile, error) {
	sortMap := make(map[string]int)
	sortMap[sortBy] = -1
	skip := numberOfItems * (page - 1)
	limit := numberOfItems
	filter := make(map[string]interface{})
	if search != "" {
		filter["$text"] = coreData.SearchOperator{Search: search}
	}
	if notIncludeUserIDList != nil && len(notIncludeUserIDList) > 0 {
		nin := make(map[string]interface{})
		nin["$nin"] = notIncludeUserIDList
		filter["objectId"] = nin
	}
	fmt.Println(filter)
	result, err := s.FindUserProfileList(filter, limit, skip, sortMap)

	return result, err
}

// FindProfileByUserIds Find profile by user IDs
func (s UserProfileServiceImpl) FindProfileByUserIds(userIds []uuid.UUID) ([]dto.UserProfile, error) {
	sortMap := make(map[string]int)
	sortMap["createdDate"] = -1

	include := make(map[string]interface{})
	include["$in"] = userIds

	filter := make(map[string]interface{})
	filter["objectId"] = include

	result, err := s.FindUserProfileList(filter, 0, 0, sortMap)

	return result, err
}

// FindByUsername find user profile by name
func (s UserProfileServiceImpl) FindByUsername(username string) (*dto.UserProfile, error) {

	filter := struct {
		Username string `json:"username"`
	}{
		Username: username,
	}
	return s.FindOneUserProfile(filter)
}

// FindByUserId find user profile by userId
func (s UserProfileServiceImpl) FindByUserId(userId uuid.UUID) (*dto.UserProfile, error) {

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: userId,
	}
	return s.FindOneUserProfile(filter)
}

// UpdateUserProfile update user profile information
func (s UserProfileServiceImpl) UpdateUserProfile(filter interface{}, data interface{}) error {

	result := <-s.UserProfileRepo.Update(userProfileCollectionName, filter, data)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateLastSeen update user profile information
func (s UserProfileServiceImpl) UpdateLastSeenNow(userId uuid.UUID) error {
	data := struct {
		LastSeen int64 `json:"lastSeen" bson:"lastSeen"`
	}{
		LastSeen: utils.UTCNowUnix(),
	}

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: userId,
	}

	updateOperator := coreData.UpdateOperator{
		Set: data,
	}

	return s.UpdateUserProfile(filter, updateOperator)
}

// UpdateUserProfileById update user profile information by user id
func (s UserProfileServiceImpl) UpdateUserProfileById(userId uuid.UUID, data *dto.UserProfile) error {
	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: userId,
	}

	updateOperator := coreData.UpdateOperator{
		Set: data,
	}

	err := s.UpdateUserProfile(filter, updateOperator)
	if err != nil {
		return err
	}
	return nil
}

// DeleteUserProfile get all user profile informaition.
func (s UserProfileServiceImpl) DeleteUserProfile(filter interface{}) error {

	result := <-s.UserProfileRepo.Delete(userProfileCollectionName, filter, true)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteManyUserProfile get all user profile informaition.
func (s UserProfileServiceImpl) DeleteManyUserProfile(filter interface{}) error {

	result := <-s.UserProfileRepo.Delete(userProfileCollectionName, filter, false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// CreateUserProfileIndex create index for user profile search.
func (s UserProfileServiceImpl) CreateUserProfileIndex(indexes map[string]interface{}) error {
	result := <-s.UserProfileRepo.CreateIndex(userProfileCollectionName, indexes)
	return result
}

// Increment increment a profile field
func (s UserProfileServiceImpl) Increment(objectId uuid.UUID, field string, value int) error {

	filter := struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: objectId,
	}

	data := make(map[string]interface{})
	data[field] = value

	incOperator := coreData.IncrementOperator{
		Inc: data,
	}
	return s.UpdateUserProfile(filter, incOperator)
}

// IncreaseFollowCount increment follow count of post
func (s UserProfileServiceImpl) IncreaseFollowCount(objectId uuid.UUID, inc int) error {
	return s.Increment(objectId, "followCount", inc)
}

// IncreaseFollowerCount increment follower count of post
func (s UserProfileServiceImpl) IncreaseFollowerCount(objectId uuid.UUID, inc int) error {
	return s.Increment(objectId, "followerCount", inc)
}
