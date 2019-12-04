package user

import "github.com/efimovad/Forums.git/internal/models"

const (
	NOT_FOUND_ERR = "Can't find user with nickname "
	NICKNAME_CONFLICT = "Data conflict by nickname "
	EMAIL_CONFLICT = "Data conflict by email "
)

type Usecase interface {
	Create(user *models.User) ([]*models.User, error)
	FindByName(nickname string) (*models.User, error)
	Edit(name string, user *models.User) error
}
