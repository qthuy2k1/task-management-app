package repository

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
)

// Add 1 user to 1 task
func (db Database) AddUserToTask(userID int, taskID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	query := `INSERT INTO user_task_details(user_id, task_id) VALUES($1, $2);`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(userID, taskID)
	if err != nil {
		return err
	}
	return nil
}

// Deletes a user from a task in the database
func (db Database) DeleteUserFromTask(userID int, taskID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	query := `DELETE FROM user_task_details WHERE user_id=$1 AND task_id=$2;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(userID, taskID)
	if err != nil {
		return err
	}
	return nil
}

// Get all the users that are assigned to the task
func (db Database) GetAllUsersAssignedToTask(taskID int) ([]models.User, error) {
	users := []models.User{}
	query := `SELECT id, name, email, role FROM user_task_details d INNER JOIN users u ON d.user_id = u.id WHERE task_id=$1;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return users, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(taskID)
	if err != nil {
		return users, err
	}
	// loop all rows and append into list
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, nil

}

// Get all the tasks that are assigned to the user
func (db Database) GetAllTaskAssignedToUser(userID int) ([]models.Task, error) {
	list := []models.Task{}
	query := `SELECT name, description FROM user_task_details d INNER JOIN tasks t ON d.task_id = t.id WHERE user_id=$1;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return list, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return list, err
	}
	// loop all rows and append into list
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.Name, &task.Description)
		if err != nil {
			return list, err
		}
		list = append(list, task)
	}
	return list, nil
}
