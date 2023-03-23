package db

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/qthuy2k1/task-management-app/models"
)

func (db Database) GetAllUsers(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.UserList, error) {
	list := &models.UserList{}
	rows, err := db.Conn.Query(`SELECT * FROM users;`)
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

func (db Database) AddUser(user *models.User) error {
	var id int
	password := user.Santize(user.Password)
	hashedPassword, err := user.Hash(password)
	if err != nil {
		return err
	}
	// insert into users table
	query := `INSERT INTO users(name, email, password, role) VALUES($1, $2, $3, $4) RETURNING id;`
	err = db.Conn.QueryRow(query, user.Name, user.Email, hashedPassword, user.Role).Scan(&id)
	if err != nil {
		return err
	}

	user.ID = id

	return nil
}

func (db Database) GetUserByID(userID int) (models.User, error) {
	user := models.User{}
	query := `SELECT * FROM users WHERE id = $1;`
	row := db.Conn.QueryRow(query, userID)
	switch err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role); err {
	case sql.ErrNoRows:
		return user, ErrNoMatch
	default:
		return user, err
	}
}

func (db Database) DeleteUser(userID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager := db.IsManager(r, tokenAuth)
	if !isManager {
		return errors.New("you are not the manager")
	}

	query := `DELETE FROM users WHERE id = $1`
	_, err := db.Conn.Exec(query, userID)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

func (db Database) UpdateUser(userID int, userData models.User) (models.User, error) {
	user := models.User{}
	password := user.Santize(user.Password)
	hashedPassword, err := user.Hash(password)
	if err != nil {
		return user, err
	}
	query := `UPDATE users SET name=$1, email=$2, password=$3, role=$4 WHERE id=$5 RETURNING *;`
	err = db.Conn.QueryRow(query, userData.Name, userData.Email, hashedPassword, userData.Role, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrNoMatch
		}
		return user, err
	}
	return user, nil
}

func (db Database) IsManager(r *http.Request, tokenAuth *jwtauth.JWTAuth) bool {
	user, err := tokenAuth.Decode(jwtauth.TokenFromCookie(r))
	if err != nil {
		return false
	}
	email, _ := user.Get("email")
	query := `SELECT * FROM users WHERE email=$1 AND role=$2;`
	result, err := db.Conn.Exec(query, email, "manager")
	if err != nil {
		return false
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return false
	}
	return true
}
func (db Database) CompareEmailAndPassword(email, password string, r *http.Request, tokenAuth *jwtauth.JWTAuth) bool {
	list, err := db.GetAllUsers(r, tokenAuth)
	if err != nil {
		return false
	}
	for _, x := range list.Users {
		if x.Email == email {
			if err = x.CheckPasswordHash(x.Password, password); err == nil {
				return true
			}
		}
	}
	return false
}
