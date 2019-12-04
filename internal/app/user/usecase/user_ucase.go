package user_ucase

import (
	"github.com/efimovad/Forums.git/internal/app/user"
	"github.com/efimovad/Forums.git/internal/models"
	"github.com/pkg/errors"
)

const (
	NOT_FOUND_ERR = "Can't find user with nickname "
	NICKNAME_CONFLICT = "Data conflict by nickname "
	EMAIL_CONFLICT = "Data conflict by email "
)


type UserUcase struct {
	repository user.Repository
}

func NewUserUsecase(r user.Repository) user.Usecase {
	return &UserUcase{
		repository: r,
	}
}

func (u * UserUcase) Create(user *models.User) ([]*models.User, error) {
	var users []*models.User
	user1, err := u.repository.FindByEmail(user.Email)
	if err == nil {
		users = append(users, user1)
	}

	user2, err := u.repository.FindByName(user.Nickname)
	if  err == nil && user1.Email != user2.Email {
		users = append(users, user2)
	}

	if len(users) != 0 {
		return users, errors.New("user already exist")
	}

	if err := u.repository.Create(user); err != nil {
		return nil, errors.Wrap(err, "repository.Create()")
	}

	return nil, nil
}

func (u *UserUcase) FindByName(nickname string) (*models.User, error) {
	myUser, err := u.repository.FindByName(nickname)
	if err != nil {
		return nil, errors.Wrap(err, "repository.FindByName()")
	}
	return myUser, nil
}

func (u *UserUcase) Edit(name string, user *models.User) error {
	currUser, err := u.repository.FindByName(name)
	if err != nil {
		return errors.New(NOT_FOUND_ERR + name)
	}

	user.Nickname = currUser.Nickname

	if currUser.Email != user.Email {
		_, err := u.repository.FindByEmail(user.Email)
		if err == nil {
			return errors.New(EMAIL_CONFLICT + user.Email)
		}
	}

	if err := u.repository.Edit(user); err != nil {
		return err
	}

	return nil
}