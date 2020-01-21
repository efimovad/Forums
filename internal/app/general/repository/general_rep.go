package general_rep

import (
	"database/sql"
	"github.com/efimovad/Forums.git/internal/app/general"
	"github.com/efimovad/Forums.git/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewGeneralRepository(db *sql.DB) general.Repository {
	return &Repository{db}
}

func (r *Repository) DropAll() error {
	if _, err := r.db.Exec("DELETE FROM posts;"); err != nil {
		return err
	}

	if _, err := r.db.Exec("DELETE FROM threads;"); err != nil {
		return err
	}

	if _, err := r.db.Exec("DELETE FROM forums;"); err != nil {
		return err
	}

	if _, err := r.db.Exec("DELETE FROM users;"); err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetStatus() (*models.ServiceInfo, error) {
	// TODO: add all tables
	info := new(models.ServiceInfo)
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users;`).
		Scan(&info.User); err != nil {
		return nil, err
	}

	if err := r.db.QueryRow(`SELECT COUNT(*) FROM threads;`).
		Scan(&info.Thread); err != nil {
		return nil, err
	}

	if err := r.db.QueryRow(`SELECT COUNT(*) FROM posts;`).
		Scan(&info.Post); err != nil {
		return nil, err
	}

	if err := r.db.QueryRow(`SELECT COUNT(*) FROM forums;`).
		Scan(&info.Forum); err != nil {
		return nil, err
	}

	return info, nil
}