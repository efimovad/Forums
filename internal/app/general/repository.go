package general

import "github.com/efimovad/Forums.git/internal/models"

type Repository interface {
	DropAll() error
	GetStatus() (*models.ServiceInfo, error)
}
