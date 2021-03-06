package user_rep

import (
	"database/sql"
	"github.com/efimovad/Forums.git/internal/app/user"
	"github.com/efimovad/Forums.git/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.Repository {
	return &Repository{db}
}

func (r *Repository) Create(user *models.User) error {
	return r.db.QueryRow(
		"INSERT INTO users (email, about, fullname, nickname) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Email,
		user.About,
		user.FullName,
		user.Nickname,
	).Scan(&user.ID)
}

func (r *Repository) FindByEmail(email string) (*models.User, error) {
	u := new(models.User)
	if err := r.db.QueryRow(
		"SELECT id, email, about, fullname, nickname FROM users WHERE LOWER(email) = LOWER($1)",
		email,
	).Scan(
		&u.ID,
		&u.Email,
		&u.About,
		&u.FullName,
		&u.Nickname,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) FindByName(nickname string) (*models.User, error) {
	u := new(models.User)
	if err := r.db.QueryRow(
		"SELECT id, email, about, fullname, nickname FROM users WHERE LOWER(nickname) = LOWER($1)",
		nickname,
	).Scan(
		&u.ID,
		&u.Email,
		&u.About,
		&u.FullName,
		&u.Nickname,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) Edit(user *models.User) error {
	return r.db.QueryRow("UPDATE users SET email = $1, about = $2, fullname = $3 "+
		"WHERE nickname = $4 RETURNING id",
		user.Email,
		user.About,
		user.FullName,
		user.Nickname,
	).Scan(&user.ID)
}