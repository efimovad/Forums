package user

import "github.com/efimovad/Forums.git/internal/models"

type Usecase interface {
	Create(user *models.User) ([]*models.User, error)
	FindByName(nickname string) (*models.User, error)
	Edit(name string, user *models.User) error
}
