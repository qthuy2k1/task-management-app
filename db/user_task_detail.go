package db

import (
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/qthuy2k1/task-management-app/models"
)

func (db Database) AddUserToTask(userID int, taskID int, r *http.Request) error {
	userToken := jwtauth.TokenFromCookie(r)

	isManager, err := db.IsManager(userToken)
	if !isManager {
		return err
	}
	query := `INSERT INTO user_task_details(user_id, task_id) VALUES($1, $2);`
	_, err = db.Conn.Exec(query, userID, taskID)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) GetAllUsersAssignedToTask(taskID int) (*models.UserList, error) {

	list := &models.UserList{}

	rows, err := db.Conn.Query(`SELECT name, email, role FROM user_task_details d INNER JOIN users u ON d.user_id = u.id WHERE task_id=$1;`, taskID)
	if err != nil {
		return list, err
	}
	// loop all rows and append into list
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Name, &user.Email, &user.Role)
		if err != nil {
			return list, err
		}
		list.Users = append(list.Users, user)
	}
	return list, nil
}

func (db Database) GetAllTaskAssignedToUser(userID int) (*models.TaskList, error) {
	list := &models.TaskList{}

	rows, err := db.Conn.Query(`SELECT name, description FROM user_task_details d INNER JOIN tasks t ON d.task_id = t.id WHERE user_id=$1;`, userID)
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
		list.Tasks = append(list.Tasks, task)
	}
	return list, nil
}
