package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/qthuy2k1/task-management-app/models"
)

// Retrieves all users from the database
func (db Database) GetAllUsers(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.UserList, error) {
	list := &models.UserList{}

	stmt, err := db.Conn.Prepare("SELECT * FROM users;")
	if err != nil {
		return list, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return list, err
	}
	// loop all rows and append into list
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role)
		if err != nil {
			return list, err
		}
		list.Users = append(list.Users, user)
	}
	return list, nil
}

// Adds a new user to the database
func (db Database) AddUser(user *models.User) error {
	// Sanitize and hash password
	password := user.Santize(user.Password)
	hashedPassword, err := user.Hash(password)
	if err != nil {
		return err
	}

	query := `INSERT INTO users(name, email, password, role) VALUES($1, $2, $3, $4) RETURNING id;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(user.Name, user.Email, hashedPassword, user.Role).Scan(&id)
	if err != nil {
		return err
	}

	// Set the ID of the user object to the newly created ID
	user.ID = id

	return nil
}

// Retrieves a user by their ID from the database
func (db Database) GetUserByID(userID int) (models.User, error) {
	user := models.User{}
	query := "SELECT * FROM users WHERE id = $1;"
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return user, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(userID)
	switch err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role); err {
	case sql.ErrNoRows:
		return user, ErrNoMatch
	default:
		return user, err
	}
}

// Retrieves a user by their email from the database
func (db Database) GetUserByEmail(userEmail string) (models.User, error) {
	user := models.User{}
	query := `SELECT * FROM users WHERE email = $1;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return user, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(userEmail)
	switch err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role); err {
	case sql.ErrNoRows:
		return user, ErrNoMatch
	default:
		return user, err
	}
}

// Deletes a user from the database, but only if the user is a manager
func (db Database) DeleteUser(userID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}

	query := `DELETE FROM users WHERE id = $1`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

// Updates a user's name and email in the database, given their ID
func (db Database) UpdateUser(userID int, userData models.User) (models.User, error) {
	user := models.User{}

	query := `UPDATE users SET name=$1, email=$2 WHERE id=$3 RETURNING id, name, email, role;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return user, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(userData.Name, userData.Email, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrNoMatch
		}
		return user, err
	}
	return user, nil
}

// Updates a user's role in the database, given their ID
func (db Database) UpdateRole(userID int, role string) (models.User, error) {
	user, err := db.GetUserByID(userID)
	if err != nil {
		return user, err
	}

	query := `UPDATE users SET role=$1 WHERE id=$2 RETURNING id, name, email, role;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return user, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(role, user.ID).Scan(&user.ID, &user.Name, &user.Email, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrNoMatch
		}
		return user, err
	}
	return user, nil
}

// Checks if a user is a manager
func (db Database) IsManager(r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (bool, error) {
	email, _ := token.Get("email")

	// Prepare a statement to retrieve the count of users with a given email and role
	stmt, err := db.Conn.Prepare(`SELECT COUNT(*) FROM users WHERE email=$1 AND role='manager'`)
	if err != nil {
		return false, err
	}

	defer stmt.Close()

	var count int
	err = stmt.QueryRow(email).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Compares the email and password entered by the user with the emails and passwords stored in the database
func (db Database) CompareEmailAndPassword(email, password string, r *http.Request, tokenAuth *jwtauth.JWTAuth) (bool, error) {
	// Get all users and assign them to list
	list, err := db.GetAllUsers(r, tokenAuth)
	if err != nil {
		return false, errors.New("cannot get list of users")
	}
	// Loop through the list of users and check if the email and password are correct
	for _, x := range list.Users {
		if x.Email == email {
			if err = x.CheckPasswordHash(x.Password, password); err == nil {
				return true, nil
			} else {
				return false, errors.New("your password is wrong")
			}
		}
	}
	return false, errors.New("your email is wrong")
}

// Changes a user's password in the database
func (db Database) ChangeUserPassword(oldPassword, newPassword string, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	// Get the token and decode it to get the email of the current user is logging in
	token, err := tokenAuth.Decode(jwtauth.TokenFromCookie(r))
	if err != nil {
		return err
	}

	user := models.User{}
	email, _ := token.Get("email")

	// Convert email from interface{} to string
	user.Email = email.(string)

	// Prepare a statement for the query to retrieve password of user
	stmt, err := db.Conn.Prepare("SELECT password FROM users WHERE email=$1")
	if err != nil {
		return err
	}

	defer stmt.Close()

	err = stmt.QueryRow(user.Email).Scan(&user.Password)
	if err != nil {
		return err
	}

	// Compare password hashed in db to the old password passed from the form value
	if err = user.CheckPasswordHash(user.Password, oldPassword); err != nil {
		return fmt.Errorf("incorrect old password")
	}

	// Hash new password
	hashedPassword, err := user.Hash(newPassword)
	if err != nil {
		return err
	}

	// Update the new hashed password for user
	stmt, err = db.Conn.Prepare("UPDATE users SET password = $1 WHERE email = $2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(hashedPassword, user.Email)
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
