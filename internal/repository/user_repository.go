package repository

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserRepository struct {
	Database *Database
}

func NewUserRepository(database *Database) *UserRepository {
	return &UserRepository{Database: database}
}

// Retrieves all users from the database
func (re *UserRepository) GetAllUsers(ctx context.Context) (models.UserSlice, error) {
	users, err := models.Users().All(ctx, re.Database.Conn)
	if err != nil {
		return users, err
	}
	return users, nil
}

// Adds a new user to the database
func (re *UserRepository) AddUser(user *models.User, ctx context.Context) error {
	err := user.Insert(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// Retrieves a user by their ID from the database
func (re *UserRepository) GetUserByID(userID int, ctx context.Context) (*models.User, error) {
	user, err := models.Users(Where("id = ?", userID)).One(ctx, re.Database.Conn)
	if err != nil {
		return user, err
	}
	return user, nil
}

// Retrieves a user by their email from the database
func (re *UserRepository) GetUserByEmail(userEmail string, ctx context.Context) (*models.User, error) {
	user, err := models.Users(Where("email = ?", userEmail)).One(ctx, re.Database.Conn)
	if err != nil {
		return user, err
	}
	return user, nil
}

// Deletes a user from the database, but only if the user is a manager
func (re *UserRepository) DeleteUser(userID int, ctx context.Context) (int64, error) {
	rowsAff, err := models.Users(Where("id = ?", userID)).DeleteAll(ctx, re.Database.Conn)
	if err != nil {
		return -1, err
	}
	return rowsAff, nil
}

// Updates a user's name and email in the database, given their ID
func (re *UserRepository) UpdateUser(user *models.User, ctx context.Context) (*models.User, error) {
	rowsAff, err := user.Update(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return user, err
	}
	if rowsAff == 0 {
		return user, ErrNoMatch
	}
	return user, nil
}

// Checks if a user is a manager
func (re *UserRepository) IsManager(ctx context.Context, email string) (bool, error) {
	count, err := models.Users(Where("email = ?", email), Where("role = ?", "manager")).Count(ctx, re.Database.Conn)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Get all users who have the role of manager
func (re *UserRepository) GetUsersManager(ctx context.Context) (models.UserSlice, error) {
	users, err := models.Users(Where("role = ?", "manager")).All(ctx, re.Database.Conn)
	if err != nil {
		return users, err
	}
	return users, err
}
