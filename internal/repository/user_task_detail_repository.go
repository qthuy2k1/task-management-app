package repository

import (
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
)

type UserTaskDetailRepository struct {
	Database *Database
}

func NewUserTaskDetailRepository(database *Database) *UserTaskDetailRepository {
	return &UserTaskDetailRepository{Database: database}
}

// Add 1 user to 1 task
func (re *UserTaskDetailRepository) AddUserToTask(userID, taskID int) error {
	query := `INSERT INTO user_task_details(user_id, task_id) VALUES($1, $2);`
	stmt, err := re.Database.Conn.Prepare(query)
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
func (re *UserTaskDetailRepository) DeleteUserFromTask(userID int, taskID int) error {
	query := `DELETE FROM user_task_details WHERE user_id=$1 AND task_id=$2;`
	stmt, err := re.Database.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(userID, taskID)
	if err != nil {
		if err == ErrNoMatch {
			return ErrNoMatch
		}
		return err
	}
	return nil
}

// Get all the users that are assigned to the task
func (re *UserTaskDetailRepository) GetAllUsersAssignedToTask(taskID int) ([]models.User, error) {
	users := []models.User{}
	query := `SELECT id, name, email, role FROM user_task_details d INNER JOIN users u ON d.user_id = u.id WHERE task_id=$1;`
	stmt, err := re.Database.Conn.Prepare(query)
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
func (re *UserTaskDetailRepository) GetAllTaskAssignedToUser(userID int) ([]models.Task, error) {
	list := []models.Task{}
	query := `SELECT id, name, description, start_date, end_date, status, author_id, created_at, updated_at, task_category_id FROM user_task_details d INNER JOIN tasks t ON d.task_id = t.id WHERE user_id=$1;`
	stmt, err := re.Database.Conn.Prepare(query)
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
		err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)
		if err != nil {
			return list, err
		}
		list = append(list, task)
	}
	return list, nil
}
