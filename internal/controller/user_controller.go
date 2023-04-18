package controller

import (
	"context"
	"errors"
	"html"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	UserRepository *repository.UserRepository
}

func NewUserController(userRepository *repository.UserRepository) *UserController {
	return &UserController{UserRepository: userRepository}
}

func (c *UserController) GetAllUsers(ctx context.Context) (models.UserSlice, error) {
	users, err := c.UserRepository.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (c *UserController) AddUser(user *models.User, ctx context.Context) error {
	// Sanitize and hash password
	password := html.EscapeString(strings.TrimSpace(user.Password))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	hashedPasswordStr := string(hashedPassword)
	user.Password = hashedPasswordStr
	if err != nil {
		return err
	}
	err = c.UserRepository.AddUser(user, ctx)
	if err != nil {
		return err
	}
	return nil
}
func (c *UserController) GetUserByID(userID int, ctx context.Context) (*models.User, error) {
	user, err := c.UserRepository.GetUserByID(userID, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (c *UserController) GetUserByEmail(userEmail string, ctx context.Context) (*models.User, error) {
	user, err := c.UserRepository.GetUserByEmail(userEmail, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (c *UserController) DeleteUser(userID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	rowsAff, err := c.UserRepository.DeleteUser(userID, ctx)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return errors.New("No row affected")
	}
	return nil
}

func (c *UserController) UpdateUser(userID int, userData models.User, ctx context.Context) (*models.User, error) {
	user, err := c.UserRepository.GetUserByID(userID, ctx)
	if err != nil {
		return user, err
	}

	user.Name = userData.Name
	user.Email = userData.Email
	userUpdated, err := c.UserRepository.UpdateUser(user, ctx)
	if err != nil {
		return user, err
	}
	return userUpdated, nil
}

func (c *UserController) UpdateRole(userID int, role string, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.User, error) {
	user, err := c.UserRepository.GetUserByID(userID, ctx)
	if err != nil {
		return user, err
	}

	user.Role = role
	userUpdated, err := c.UserRepository.UpdateUser(user, ctx)
	if err != nil {
		return user, err
	}
	return userUpdated, nil
}

// Checks if a user is a manager
func (c *UserController) IsManager(ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	token, err := tokenAuth.Decode(jwtauth.TokenFromCookie(r))
	if err != nil {
		return err
	}

	email, _ := token.Get("email")

	// Convert email from interface{} to string
	emailStr, ok := email.(string)
	if !ok {
		return errors.New("cannot convert email from interface to string")
	}

	isManager, err := c.UserRepository.IsManager(ctx, emailStr)

	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}

	return nil
}

func (c *UserController) CompareEmailAndPassword(email, password string, ctx context.Context) (bool, error) {
	users, err := c.UserRepository.GetAllUsers(ctx)
	if err != nil {
		return false, err
	}
	for _, x := range users {
		if x.Email == email {
			if err = bcrypt.CompareHashAndPassword([]byte(x.Password), []byte(password)); err == nil {
				return true, nil
			} else {
				return false, errors.New("your password is wrong")
			}
		}
	}
	return false, errors.New("your email is wrong")
}

func (c *UserController) ChangeUserPassword(oldPassword, newPassword, email string, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	newPassword = html.EscapeString(strings.TrimSpace(newPassword))
	oldPassword = html.EscapeString(strings.TrimSpace(oldPassword))

	user, err := c.UserRepository.GetUserByEmail(email, ctx)
	if err != nil {
		return err
	}

	// Compare password hashed in db to the old password got from the form value
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("incorrect old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	_, err = c.UserRepository.UpdateUser(user, ctx)
	if err != nil {
		return err
	}
	return nil
}
