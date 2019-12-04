package general_ucase

import (
	"github.com/efimovad/Forums.git/internal/app/general"
	"github.com/efimovad/Forums.git/internal/models"
)

type Usecase struct {
	repository general.Repository
}

func NewGeneralUsecase(r general.Repository) general.Usecase {
	return &Usecase{
		repository: r,
	}
}

func (u *Usecase) GetStatus() (*models.ServiceInfo, error) {
	return u.repository.GetStatus()
}

func (u *Usecase) DropAll() error {
	return u.repository.DropAll()
}
