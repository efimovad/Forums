package user

import "github.com/efimovad/Forums.git/internal/models"

type Repository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByName(nickname string) (*models.User, error)
	Edit(user *models.User) error
}
