package db

import (
	"database/sql"

	"github.com/qthuy2k1/task-management-app/models"
)

func (db Database) GetAllUsers() (*models.UserList, error) {
	list := &models.UserList{}
	rows, err := db.Conn.Query(`SELECT * FROM users;`)
	if err != nil {
		return list, err
	}
	// loop all rows and append into list
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.Token)
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
	query := `INSERT INTO users(name, email, password, role, token) VALUES($1, $2, $3, $4, $5) RETURNING id;`
	err = db.Conn.QueryRow(query, user.Name, user.Email, hashedPassword, user.Role, user.Token).Scan(&id)
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
	switch err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.Token); err {
	case sql.ErrNoRows:
		return user, ErrNoMatch
	default:
		return user, err
	}
}

func (db Database) DeleteUser(userID int) error {
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
	query := `UPDATE users SET name=$1, email=$2, password=$3, role=$4 token=$5 WHERE id=$6 RETURNING *;`
	err = db.Conn.QueryRow(query, userData.Name, userData.Email, hashedPassword, userData.Role, userData.Token, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrNoMatch
		}
		return user, err
	}
	return user, nil
}

func (db Database) IsManager(token string) (bool, error) {
	query := `SELECT * FROM users WHERE token=$1, role=$2;`
	_, err := db.Conn.Exec(query, token, "manager")
	switch err {
	case sql.ErrNoRows:
		return false, ErrNoMatch
	default:
		return true, nil
	}
}
