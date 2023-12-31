package service

import (
	"buddle-server/model"
	"buddle-server/repository"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	SignUp(c context.Context, user *model.User) error
	SignIn(c context.Context, user *model.User) (*model.Response, error)
}

type userService struct {
	repo repository.Repository
}

func NewUserService(repo repository.Repository) (UserService, error) {
	if repo == nil {
		return nil, errors.New("repository is nil")
	}
	return &userService{repo: repo}, nil
}

func (u userService) SignUp(c context.Context, user *model.User) error {
	switch {
	case c == nil:
		return errors.New("nil context")
	case user == nil:
		return errors.New("user is nil")
	}

	oriUser, err := u.repo.User().GetUserByID(c, user.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "failed to check user for sign up")
	}

	if oriUser != nil && oriUser.UserSeq != 0 {
		return fmt.Errorf("user(%d) is duplicated", oriUser.UserSeq)
	}

	// 비밀번호를 bycrypt 라이브러리로 해싱 처리
	hashpw, err := HashPassword(user.Password)
	if err != nil {
		return errors.New("failed to encrypt password")
	}
	user.Password = hashpw

	if err := u.repo.User().Create(c, user); err != nil {
		return errors.Wrap(err, "failed to sign up user")
	}

	return nil
}

func (u userService) SignIn(c context.Context, user *model.User) (*model.Response, error) {
	switch {
	case c == nil:
		return model.SimpleFail(), errors.New("nil context")
	case user == nil:
		return model.SimpleFail(), errors.New("user is nil")
	}

	inputId := user.Id
	inputPassword := user.Password

	user, err := u.repo.User().GetUserByID(c, user.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("invaild user [ id = %s ] err : %+v", inputId, err)
			return &model.Response{
				Success:   false,
				Message:   "아이디가 존재 하지 않습니다.",
				ErrorCode: model.ResponseErrorCodeUserIDNotExist,
			}, nil
		}
		return model.SimpleFail(), errors.Wrap(err, "failed to check user for sign up")
	}

	if !CheckPasswordHash(user.Password, inputPassword) {
		return &model.Response{
			Success:   false,
			Message:   "비밀번호가 일지하지 않습니다.",
			ErrorCode: model.ResponseErrorCodeInvalidUserPwd,
		}, nil
	}

	return model.SimpleSuccess(), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(hashVal, userPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashVal), []byte(userPw))
	if err != nil {
		return false
	} else {
		return true
	}
}
