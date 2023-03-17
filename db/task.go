package db

import (
	"database/sql"

	"time"

	"github.com/qthuy2k1/task-management-app/models"
)

func (db Database) GetAllTasks() (*models.TaskList, error) {
	list := &models.TaskList{}
	rows, err := db.Conn.Query(`SELECT * FROM tasks;`)
	if err != nil {
		return list, err
	}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.UserID, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)
		if err != nil {
			return list, err
		}
		list.Tasks = append(list.Tasks, task)
	}
	return list, nil
}
func (db Database) AddTask(task *models.Task) error {
	var id int
	var createdAt time.Time
	query := `insert into tasks(name, description, start_date, end_date, status, user_id, author_id, task_category_id) values($1, $2, $3, $4, $5, $6, $7, $8) returning id, created_at;`
	err := db.Conn.QueryRow(query, task.Name, task.Description, task.StartDate, task.EndDate, task.Status, task.UserID, task.AuthorID, task.TaskCategoryID).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	task.ID = id
	task.CreatedAt = createdAt
	return nil
}
func (db Database) GetTaskById(taskId int) (models.Task, error) {
	task := models.Task{}
	query := `SELECT * FROM tasks WHERE id = $1;`
	row := db.Conn.QueryRow(query, taskId)
	switch err := row.Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.UserID, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID); err {
	case sql.ErrNoRows:
		return task, ErrNoMatch
	default:
		return task, err
	}
}
func (db Database) DeleteTask(taskId int) error {
	query := `DELETE FROM tasks WHERE id = $1;`
	_, err := db.Conn.Exec(query, taskId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}
func (db Database) UpdateTask(taskId int, taskData models.Task) (models.Task, error) {
	task := models.Task{}
	query := `UPDATE tasks SET name=$1, description=$2, start_date=$3, end_date=$4, status=$5, user_id=$6, author_id=$7, updated_at=$8, task_category_id=$9 WHERE id=$10 RETURNING *;`
	err := db.Conn.QueryRow(query, taskData.Name, taskData.Description, taskData.StartDate, taskData.EndDate, taskData.Status, taskData.UserID, taskData.AuthorID, time.Now(), taskData.TaskCategoryID, taskId).Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.UserID, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, ErrNoMatch
		}
		return task, err
	}
	return task, nil
}
