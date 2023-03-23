package db

import (
	"database/sql"
	"errors"
	"net/http"

	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/qthuy2k1/task-management-app/models"
)

func (db Database) GetAllTasks(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.TaskList, error) {
	list := &models.TaskList{}
	rows, err := db.Conn.Query(`SELECT * FROM tasks;`)
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
		list.Tasks = append(list.Tasks, task)
	}
	return list, nil
}
func (db Database) AddTask(task *models.Task, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager := db.IsManager(r, tokenAuth)

	if !isManager {
		return errors.New("you are not the manager")
	}

	var id int
	var createdAt time.Time

	// insert into tasks table
	query := `insert into tasks(name, description, start_date, end_date, status, author_id, task_category_id) values($1, $2, $3, $4, $5, $6, $7) returning id, created_at;`
	err := db.Conn.QueryRow(query, task.Name, task.Description, task.StartDate, task.EndDate, task.Status, task.AuthorID, task.TaskCategoryID).Scan(&id, &createdAt)
	if err != nil {
		return err
	}

	task.ID = id
	task.CreatedAt = createdAt
	return nil
}
func (db Database) GetTaskByID(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.Task, error) {
	task := models.Task{}
	query := `SELECT * FROM tasks WHERE id = $1;`
	row := db.Conn.QueryRow(query, taskID)
	switch err := row.Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID); err {
	case sql.ErrNoRows:
		return task, ErrNoMatch
	default:
		return task, err
	}
}
func (db Database) DeleteTask(taskId int, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager := db.IsManager(r, tokenAuth)

	if !isManager {
		return errors.New("you are not the manager")
	}

	query := `DELETE FROM tasks WHERE id = $1;`
	_, err := db.Conn.Exec(query, taskId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}
func (db Database) UpdateTask(taskID int, taskData models.Task, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.Task, error) {
	task := models.Task{}

	isManager := db.IsManager(r, tokenAuth)
	query := ""

	if isManager {
		query = `UPDATE tasks SET name=$1, description=$2, start_date=$3, end_date=$4, status=$5, author_id=$6, updated_at=$7, task_category_id=$8 WHERE id=$9 RETURNING *;`
	} else {
		query = `UPDATE tasks SET name=$1, description=$2, start_date=$3, end_date=$4, status=$5, author_id=$6, updated_at=$7, task_category_id=$8 WHERE id=$9 AND status NOT IN ("Lock") RETURNING *;`
	}
	err := db.Conn.QueryRow(query, taskData.Name, taskData.Description, taskData.StartDate, taskData.EndDate, taskData.Status, taskData.AuthorID, time.Now(), taskData.TaskCategoryID, taskID).Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, ErrNoMatch
		}
		return task, err
	}
	return task, nil
}

func (db Database) LockTask(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager := db.IsManager(r, tokenAuth)
	if !isManager {
		return errors.New("you are not the manager")
	}
	task := models.Task{}
	query := `UPDATE tasks SET status=$1 WHERE id=$2 RETURNING *;`
	err := db.Conn.QueryRow(query, "Lock", taskID).Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNoMatch
		}
		return err
	}
	return nil
}

func (db Database) UnLockTask(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager := db.IsManager(r, tokenAuth)
	if !isManager {
		return errors.New("you are not the manager")
	}
	task := models.Task{}
	query := `UPDATE tasks SET status=$1 WHERE id=$2 RETURNING *;`
	err := db.Conn.QueryRow(query, "Working", taskID).Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNoMatch
		}
		return err
	}
	return nil
}
