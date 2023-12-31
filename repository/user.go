package repository

import (
	"buddle-server/internal/db"
	"buddle-server/model"
	"context"
	"github.com/pkg/errors"
	"time"
)

type UserRepository interface {
	Create(c context.Context, user *model.User) error
    GetUserByID(c context.Context, id string) (*model.User, error)
}

type userRepository struct {

}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (u userRepository) Create(c context.Context, user *model.User) error {
	switch {
	case c == nil:
		return errors.New("nil context")
	case user == nil:
		return errors.New("user is nil")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return errors.Wrap(err, "failed to get db connection")
	}

	toDay := time.Now()
	user.Regdate = toDay

	if err := conn.Create(&user).Error; err != nil {
		return errors.Wrapf(err, "failed to create user [ user = %+v ]", user)
	}

	return nil
}

func (u userRepository) GetUserByID(c context.Context, id string) (*model.User, error) {
	switch {
	case c == nil:
		return nil, errors.New("nil context")
	case id == "":
		return nil, errors.New("id is empty")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	user := new(model.User)
	if err := conn.Where("id = ?", id).Take(&user).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get user by id(%s)", id)
	}

	return user, nil
}
