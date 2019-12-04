package general

import "github.com/efimovad/Forums.git/internal/models"

type Usecase interface {
	DropAll() error
	GetStatus() (*models.ServiceInfo, error)
}
