package service

import (
	"fmt"

	uuid "github.com/gofrs/uuid"
	coreConfig "github.com/red-gold/telar-core/config"
	repo "github.com/red-gold/telar-core/data"
	"github.com/red-gold/telar-core/data/mongodb"
	mongoRepo "github.com/red-gold/telar-core/data/mongodb"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	authConfig "github.com/red-gold/telar-web/micros/auth/config"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
)

// UserVerificationService handlers with injected dependencies
type UserVerificationServiceImpl struct {
	UserVerificationRepo repo.Repository
}

type EmailVerificationToken struct {
	UserId          uuid.UUID
	HtmlTmplPath    string
	Username        string
	RemoteIpAddress string
	EmailTo         string
	EmailSubject    string
	FullName        string
	UserPassword    string
}

type PhoneVerificationToken struct {
	UserId          uuid.UUID
	UserEmail       string
	Username        string
	RemoteIpAddress string
	PhoneNumber     string
	FullName        string
	UserPassword    string
}

type MetaVerificationTokenClaim struct {
	UserId          uuid.UUID             `json:"userId"`
	VerifyId        uuid.UUID             `json:"verifyId"`
	RemoteIpAddress string                `json:"remoteIpAddress"`
	Mode            constants.TokenConst  `json:"mode"`
	VerifyType      constants.VerifyConst `json:"verifyType"`
	Fullname        string                `json:"fullName"`
	Email           string                `json:"email"`
	Password        string                `json:"password"`
}

// NewUserVerificationService initializes UserVerificationService's dependencies and create new UserVerificationService struct
func NewUserVerificationService(db interface{}) (UserVerificationService, error) {

	userVerificationService := &UserVerificationServiceImpl{}

	switch *coreConfig.AppConfig.DBType {
	case coreConfig.DB_MONGO:

		mongodb := db.(mongodb.MongoDatabase)
		userVerificationService.UserVerificationRepo = mongoRepo.NewDataRepositoryMongo(mongodb)

	}
	if userVerificationService.UserVerificationRepo == nil {
		fmt.Printf("userVerificationService.UserVerificationRepo is nil! \n")
	}
	return userVerificationService, nil
}

// SaveUserVerification save user authentication informaition
func (s UserVerificationServiceImpl) SaveUserVerification(userVerification *dto.UserVerification) error {

	if userVerification.ObjectId == uuid.Nil {
		var uuidErr error
		userVerification.ObjectId, uuidErr = uuid.NewV4()
		if uuidErr != nil {
			return uuidErr
		}
	}

	if userVerification.CreatedDate == 0 {
		userVerification.CreatedDate = utils.UTCNowUnix()
	}

	result := <-s.UserVerificationRepo.Save(userVerificationCollectionName, userVerification)

	return result.Error
}

// FindOneUserVerification get all user authentication informaition
func (s UserVerificationServiceImpl) FindOneUserVerification(filter interface{}) (*dto.UserVerification, error) {

	result := <-s.UserVerificationRepo.FindOne(userVerificationCollectionName, filter)
	if result.Error() != nil {
		return nil, result.Error()
	}

	var userVerificationResult dto.UserVerification
	errDecode := result.Decode(&userVerificationResult)
	if errDecode != nil {
		return nil, fmt.Errorf("Error docoding on dto.UserVerification")
	}
	return &userVerificationResult, nil
}

// FindUserVerificationList get all user authentication informaition
func (s UserVerificationServiceImpl) FindUserVerificationList(filter interface{}, limit int64, skip int64, sort map[string]int) ([]dto.UserVerification, error) {

	result := <-s.UserVerificationRepo.Find(userVerificationCollectionName, filter, limit, skip, sort)
	defer result.Close()
	if result.Error() != nil {
		return nil, result.Error()
	}
	var userVerificationList []dto.UserVerification
	for result.Next() {
		var userVerification dto.UserVerification
		errDecode := result.Decode(&userVerification)
		if errDecode != nil {
			return nil, fmt.Errorf("Error docoding on dto.UserVerification")
		}
		userVerificationList = append(userVerificationList, userVerification)
	}

	return userVerificationList, nil
}

// FindByVerifyId find user verification record by verify id
func (s UserVerificationServiceImpl) FindByVerifyId(verifyId uuid.UUID) (*dto.UserVerification, error) {

	return s.FindOneUserVerification(struct {
		ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
	}{
		ObjectId: verifyId,
	})
}

// UpdateUserVerification get all user authentication informaition
func (s UserVerificationServiceImpl) UpdateUserVerification(filter interface{}, data interface{}) error {

	result := <-s.UserVerificationRepo.Update(userVerificationCollectionName, filter, data)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteUserVerification get all user authentication informaition
func (s UserVerificationServiceImpl) DeleteUserVerification(filter interface{}) error {

	result := <-s.UserVerificationRepo.Delete(userVerificationCollectionName, filter, true)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteManyUserVerification get all user authentication informaition
func (s UserVerificationServiceImpl) DeleteManyUserVerification(filter interface{}) error {

	result := <-s.UserVerificationRepo.Delete(userVerificationCollectionName, filter, false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteManyUserVerification get all user authentication informaition
func (s UserVerificationServiceImpl) FindByUserId(userId uuid.UUID) (*dto.UserVerification, error) {

	filter := struct {
		UserId uuid.UUID `json:"userId"`
	}{
		UserId: userId,
	}
	return s.FindOneUserVerification(filter)
}

// VerifyUserByCode verify user by verification code
func (s UserVerificationServiceImpl) VerifyUserByCode(userId uuid.UUID, verifyId uuid.UUID, remoteIpAddress string, code string, target string) (bool, error) {
	userVerification, findErr := s.FindByVerifyId(verifyId)
	if findErr != nil {
		fmt.Println(findErr.Error())
		return false, fmt.Errorf("verifyUserByCode/invalidVerifyId")
	}

	if userVerification.RemoteIpAddress != remoteIpAddress {
		return false, fmt.Errorf("verifyUserByCode/differentRemoteAddress")
	}

	newCounter := userVerification.Counter + 1
	userVerification.Counter = newCounter
	if newCounter > numberOfVerifyRequest {
		return false, fmt.Errorf("verifyUserByCode/exceedRequestsLimits")
	}
	if userVerification.IsVerified {
		return false, fmt.Errorf("verifyUserByCode/alreadyVerified")
	}

	if userVerification.Target != target {
		return false, fmt.Errorf("verifyUserByCode/differentTarget %s : %s", userVerification.Target, target)
	}
	filter := struct {
		ObjectId uuid.UUID `json:"objectId"`
	}{
		ObjectId: verifyId,
	}
	fmt.Printf("\nCode: %s , User code: %s\n", userVerification.Code, code)
	if userVerification.Code != code {
		userVerification.LastUpdated = utils.UTCNowUnix()

		updateData := struct {
			Set interface{} `json:"$set" bson:"$set"`
		}{
			Set: struct {
				LastUpdated int64 `json:"last_updated"`
				Counter     int64 `json:"counter"`
			}{
				LastUpdated: userVerification.LastUpdated,
				Counter:     userVerification.Counter,
			},
		}

		err := s.UpdateUserVerification(filter, updateData)
		if err != nil {
			fmt.Println(findErr.Error())
			return false, fmt.Errorf("createCodeVerification/updateVerificationCode")
		}
		return false, fmt.Errorf("createCodeVerification/wrongPinCod")
	}

	if utils.IsTimeExpired(userVerification.CreatedDate, expireTimeOffset) {
		return false, fmt.Errorf("verifyUserByCode/codeExpired")
	}

	// Set verify status true
	updateData := struct {
		Set interface{} `json:"$set" bson:"$set"`
	}{
		Set: struct {
			LastUpdated int64 `json:"last_updated"`
			Counter     int64 `json:"counter"`
			IsVerified  bool  `json:"isVerified"`
		}{
			LastUpdated: userVerification.LastUpdated,
			Counter:     userVerification.Counter,
			IsVerified:  true,
		},
	}
	err := s.UpdateUserVerification(filter, updateData)
	if err != nil {
		fmt.Println(findErr.Error())
		return false, fmt.Errorf("createCodeVerification/updateVerificationCode")
	}

	return true, nil
}

// CreateEmailVerficationToken Create email verification token
func (s UserVerificationServiceImpl) CreateEmailVerficationToken(input EmailVerificationToken,
	coreConfig *coreConfig.Configuration) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	// Send email

	email := utils.NewEmail(*coreConfig.RefEmail, *coreConfig.RefEmailPass, *coreConfig.SmtpEmail)
	emailReq := utils.NewEmailRequest([]string{input.EmailTo}, input.EmailSubject, input.HtmlTmplPath)

	code := utils.GenerateDigits(6)
	emailResStatus, emailResErr := email.SendEmail(emailReq, input.HtmlTmplPath, struct {
		Name      string
		AppName   string
		AppURL    string
		Code      string
		OrgName   string
		OrgAvatar string
	}{
		Name:      input.FullName,
		AppName:   *coreConfig.AppName,
		AppURL:    authConfig.AuthConfig.WebURL,
		Code:      code,
		OrgName:   *coreConfig.OrgName,
		OrgAvatar: *coreConfig.OrgAvatar,
	})

	if emailResErr != nil {
		return "", fmt.Errorf("Error happened in sending email error: %s", emailResErr.Error())
	}
	if !emailResStatus {
		return "", fmt.Errorf("Email response status is false! ")
	}
	fmt.Println("Email has been sent!")
	verifyId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	userVerification := &dto.UserVerification{
		ObjectId:        verifyId,
		UserId:          input.UserId,
		Code:            code,
		Target:          input.EmailTo,
		TargetType:      constants.EmailVerifyConst,
		Counter:         1,
		RemoteIpAddress: input.RemoteIpAddress,
	}
	saveErr := s.SaveUserVerification(userVerification)
	if saveErr != nil {
		return "", saveErr
	}

	metaToken := MetaVerificationTokenClaim{
		UserId:          input.UserId,
		VerifyId:        verifyId,
		RemoteIpAddress: input.RemoteIpAddress,
		Mode:            constants.RegisterationTokenConst,
		VerifyType:      constants.EmailVerifyConst,
		Fullname:        input.FullName,
		Email:           input.EmailTo,
		Password:        input.UserPassword,
	}

	return utils.GenerateJWTToken([]byte(*coreConfig.PrivateKey), utils.TokenClaims{
		Claim: metaToken,
	}, 1)
}

// CreatePhoneVerficationToken Create phone verification token
func (s UserVerificationServiceImpl) CreatePhoneVerficationToken(input PhoneVerificationToken,
	coreConfig *coreConfig.Configuration) (string, error) {

	// Send SMS

	phone, phoneErr := utils.NewPhone(*coreConfig.PhoneAuthToken, *coreConfig.PhoneAuthId, *coreConfig.PhoneSourceNumber)
	if phoneErr != nil {
		return "", phoneErr
	}
	code := utils.GenerateDigits(6)
	_, SmsErr := phone.SendSms(input.PhoneNumber, code)

	if SmsErr != nil {
		return "", fmt.Errorf("Error happened in sending sms error: %s", SmsErr.Error())
	}

	verifyId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	userVerification := &dto.UserVerification{
		ObjectId:        verifyId,
		UserId:          input.UserId,
		Code:            code,
		Target:          input.PhoneNumber,
		TargetType:      constants.PhoneVerifyConst,
		Counter:         1,
		RemoteIpAddress: input.RemoteIpAddress,
	}
	saveErr := s.SaveUserVerification(userVerification)
	if saveErr != nil {
		return "", saveErr
	}

	metaToken := MetaVerificationTokenClaim{
		UserId:          input.UserId,
		VerifyId:        verifyId,
		RemoteIpAddress: input.RemoteIpAddress,
		Mode:            constants.RegisterationTokenConst,
		VerifyType:      constants.EmailVerifyConst,
		Fullname:        input.FullName,
		Email:           input.UserEmail,
		Password:        input.UserPassword,
	}

	// Generate JWT token
	return utils.GenerateJWTToken([]byte(*coreConfig.PrivateKey), utils.TokenClaims{
		Claim: metaToken,
	}, 1)
}
