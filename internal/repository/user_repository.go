package repository

import (
	"context"

	"github.com/slmbngl/OrderAplication/internal/adapters/db"
	"github.com/slmbngl/OrderAplication/internal/models"
)

type UserRepository interface {
	Create(user *models.User) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
}
type userRepo struct{}

func NewUserRepository() UserRepository {
	return &userRepo{}
}

func (r *userRepo) Create(user *models.User) (*models.User, error) {
	var userID int
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, is_active) VALUES ($1, $2, $3) RETURNING id`,
		user.Username, user.PasswordHash, user.IsActive).Scan(&userID)

	if err != nil {
		return nil, err
	}

	user.ID = userID
	return user, nil
}

func (r *userRepo) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, username, password_hash, is_active, created_at FROM users WHERE username = $1`,
		username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsActive, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
