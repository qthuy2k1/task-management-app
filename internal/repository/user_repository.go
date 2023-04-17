package repository

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
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
	// Sanitize and hash password
	password := html.EscapeString(strings.TrimSpace(user.Password))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	hashedPasswordStr := string(hashedPassword)
	user.Password = hashedPasswordStr
	if err != nil {
		return err
	}

	fmt.Println(user)
	err = user.Insert(ctx, re.Database.Conn, boil.Infer())
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
func (re *UserRepository) UpdateUser(userID int, userData models.User, ctx context.Context) (*models.User, error) {
	user, err := models.Users(Where("id = ?", userID)).One(ctx, re.Database.Conn)
	if err != nil {
		return user, err
	}
	user.Name = userData.Name
	user.Email = userData.Email
	rowsAff, err := user.Update(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return user, err
	}
	if rowsAff == 0 {
		return user, ErrNoMatch
	}
	return user, nil
}

// Updates a user's role in the database, given their ID
func (re *UserRepository) UpdateRole(userID int, role string, ctx context.Context) (*models.User, error) {
	user, err := models.Users(Where("id = ?", userID)).One(ctx, re.Database.Conn)
	if err != nil {
		return user, err
	}
	user.Role = role
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

// Compares the email and password entered by the user with the emails and passwords stored in the database
func (re *UserRepository) CompareEmailAndPassword(email, password string, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth) (bool, error) {
	// Get all users and assign them to list
	users, err := models.Users().All(ctx, re.Database.Conn)
	if err != nil {
		return false, errors.New("cannot get list of users")
	}
	// Loop through the list of users and check if the email and password are correct
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

// Changes a user's password in the database
func (re *UserRepository) ChangeUserPassword(oldPassword, newPassword string, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	// Get the token and decode it to get the email of the current user is logging in
	token, err := tokenAuth.Decode(jwtauth.TokenFromCookie(r))
	if err != nil {
		return err
	}

	email, _ := token.Get("email")

	// Convert email from interface{} to string
	email = email.(string)

	user, err := models.Users(Where("email = ?", email)).One(ctx, re.Database.Conn)
	if err != nil {
		return err
	}

	// Compare password hashed in db to the old password passed from the form value
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return fmt.Errorf("incorrect old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	_, err = user.Update(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

// Validates that an email address is in a valid format
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Define a regular expression for validating email addresses
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Use the MatchString method to check if the email address matches the regular expression
	return emailRegex.MatchString(email)
}

// Validates that a password meets the minimum requirements
func IsValidPassword(password string) bool {
	if password == "" {
		return false
	}
	// Check if the password is at least 6 characters long
	return len(password) >= 6
}
