package db

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/qthuy2k1/task-management-app/models"
)

func (db Database) GetAllTaskCategories(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.TaskCategoryList, error) {
	list := &models.TaskCategoryList{}
	rows, err := db.Conn.Query(`SELECT * FROM task_categories;`)
	if err != nil {
		return list, err
	}
	// loop all rows and append into list
	for rows.Next() {
		var taskCategory models.TaskCategory
		err := rows.Scan(&taskCategory.ID, &taskCategory.Name)
		if err != nil {
			return list, err
		}
		list.TaskCategories = append(list.TaskCategories, taskCategory)
	}
	return list, nil
}

func (db Database) AddTaskCategory(taskCategory *models.TaskCategory, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager := db.IsManager(r, tokenAuth)
	if !isManager {
		return errors.New("you are not the manager")
	}
	var id int
	// insert into taskCategorys table
	query := `INSERT INTO task_categories(name) VALUES($1) RETURNING id;`
	err := db.Conn.QueryRow(query, taskCategory.Name).Scan(&id)
	if err != nil {
		return err
	}
	taskCategory.ID = id
	return nil
}

func (db Database) GetTaskCategoryByID(taskCategoryID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.TaskCategory, error) {
	taskCategory := models.TaskCategory{}
	query := `SELECT * FROM task_categories WHERE id = $1;`
	row := db.Conn.QueryRow(query, taskCategoryID)
	switch err := row.Scan(&taskCategory.ID, &taskCategory.Name); err {
	case sql.ErrNoRows:
		return taskCategory, ErrNoMatch
	default:
		return taskCategory, err
	}
}

func (db Database) DeleteTaskCategory(taskCategoryID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager := db.IsManager(r, tokenAuth)
	if !isManager {
		return errors.New("you are not the manager")
	}
	query := `DELETE FROM task_categories WHERE id = $1`
	_, err := db.Conn.Exec(query, taskCategoryID)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

func (db Database) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.TaskCategory, error) {
	taskCategory := models.TaskCategory{}
	isManager := db.IsManager(r, tokenAuth)
	if !isManager {
		return taskCategory, errors.New("you are not the manager")
	}
	query := `UPDATE task_categories SET name=$1 WHERE id=$2 RETURNING *;`
	err := db.Conn.QueryRow(query, taskCategoryData.Name, taskCategoryData.ID).Scan(&taskCategory.ID, &taskCategory.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return taskCategory, ErrNoMatch
		}
		return taskCategory, err
	}
	return taskCategory, nil
}
