package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
)

type UserController struct {
	UserRepository *repository.UserRepository
}

func NewUserController(userRepository *repository.UserRepository) *UserController {
	return &UserController{UserRepository: userRepository}
}

func (re *UserController) GetAllUsers(ctx context.Context) (models.UserSlice, error) {
	users, err := re.UserRepository.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (re *UserController) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	user, err := re.UserRepository.GetUserByID(userID, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (re *UserController) GetUserByEmail(ctx context.Context, userEmail string) (*models.User, error) {
	user, err := re.UserRepository.GetUserByEmail(userEmail, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (re *UserController) DeleteUser(userID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := re.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return err
	}
	rowsAff, err := re.UserRepository.DeleteUser(userID, ctx)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return errors.New("No row affected")
	}
	return nil
}

// Checks if a user is a manager
func (re *UserController) IsManager(ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (bool, error) {
	email, _ := token.Get("email")

	isManager, err := re.UserRepository.IsManager(ctx, email.(string))

	if err != nil {
		return isManager, err
	}
	if !isManager {
		return isManager, errors.New("you are not the manager")
	}

	return isManager, nil
}
