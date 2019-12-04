package user_rep

import (
	"database/sql"
	"github.com/efimovad/Forums.git/internal/app/user"
	"github.com/efimovad/Forums.git/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.Repository {
	return &UserRepository{db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.QueryRow(
		"INSERT INTO users (email, about, fullname, nickname) VALUES ($1, $2, $3, $4) RETURNING nickname",
		user.Email,
		user.About,
		user.FullName,
		user.Nickname,
	).Scan(&user.Nickname)
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	u := new(models.User)
	if err := r.db.QueryRow(
		"SELECT email, about, fullname, nickname FROM users WHERE email = $1",
		email,
	).Scan(
		&u.Email,
		&u.About,
		&u.FullName,
		&u.Nickname,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindByName(nickname string) (*models.User, error) {
	u := new(models.User)
	if err := r.db.QueryRow(
		"SELECT email, about, fullname, nickname FROM users WHERE nickname = $1",
		nickname,
	).Scan(
		&u.Email,
		&u.About,
		&u.FullName,
		&u.Nickname,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Edit(user *models.User) error {
	return r.db.QueryRow("UPDATE users SET email = $1, about = $2, fullname = $3 "+
		"WHERE nickname = $4 RETURNING nickname",
		user.Email,
		user.About,
		user.FullName,
		user.Nickname,
	).Scan(&user.Nickname)
}
