package repository

import (
	"context"

	"github.com/slmbngl/OrderAplication/internal/adapters/db"
	"github.com/slmbngl/OrderAplication/internal/models"
)

type UserRepository interface {
	Create(user *models.User) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUserRole(userID int, role string) error
	GetByID(userID int) (*models.GetMeResponseReq, error) // Optional: Get user by ID
}
type userRepo struct{}

func NewUserRepository() UserRepository {
	return &userRepo{}
}

func (r *userRepo) Create(user *models.User) (*models.User, error) {
	var userID int
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, is_active, role) VALUES ($1, $2, $3, $4) RETURNING id`,
		user.Username, user.PasswordHash, user.IsActive, user.Role).Scan(&userID)

	if err != nil {
		return nil, err
	}
	// Set the ID and default role if not specified
	if user.Role == "" {
		user.Role = "user" // Default role if not specified
	}

	user.ID = userID
	return user, nil
}

func (r *userRepo) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, username, password_hash, is_active, role, created_at FROM users WHERE username = $1`,
		username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAllUsers retrieves all users from the database
func (r *userRepo) GetAllUsers() ([]models.User, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, username, is_active, role, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.IsActive, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		// Don't include password hash in the response
		user.PasswordHash = ""
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserRole updates a user's role in the database
func (r *userRepo) UpdateUserRole(userID int, role string) error {
	result, err := db.Pool.Exec(context.Background(),
		`UPDATE users SET role = $1 WHERE id = $2`,
		role, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return &UserNotFoundError{UserID: userID}
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepo) GetByID(userID int) (*models.GetMeResponseReq, error) {
	var user models.GetMeResponseReq
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, username, is_active, created_at FROM users WHERE id = $1`,
		userID).Scan(&user.ID, &user.Username, &user.IsActive, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Custom error type for user not found
type UserNotFoundError struct {
	UserID int
}

func (e *UserNotFoundError) Error() string {
	return "user not found"
}
