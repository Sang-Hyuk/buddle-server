package model

import (
	"errors"
	"fmt"
	"time"
)

type UserType int

const (
	UserTypeAdmin UserType = 0 // 관리자
)

func (t UserType) Validate() error {
	switch t {
	case UserTypeAdmin:
		return nil
	}

	return fmt.Errorf("user_type(%d) is invalid", t)
}

type User struct {
	UserSeq  int64     `json:"user_seq,omitempty" gorm:"Column:user_seq;PRIMARY_KEY"`
	Id       string    `json:"id,omitempty" gorm:"Column:id"`
	Password string    `json:"password,omitempty" gorm:"Column:password"`
	UserType UserType  `json:"user_type,omitempty" gorm:"Column:user_type"`
	Name     string    `json:"name,omitempty" gorm:"Column:name"`
	Phone    string    `json:"phone,omitempty" gorm:"Column:phone"`
	Birth    string    `json:"birth,omitempty" gorm:"Column:birth"`
	Regdate  time.Time `json:"regdate" gorm:"Column:regdate"`
	Modified time.Time `json:"modified" gorm:"Column:modified"`
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) SignUpCheck() (err error) {
	switch {
	case u.Id == "":
		err = errors.New("id is empty")
	case u.Password == "":
		err = errors.New("password is empty")
	}

	return
}