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
	query := "SELECT * FROM task_categories"
	stmt, err := db.Conn.Prepare(query)
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
	isManager, err := db.IsManager(r, tokenAuth)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	var id int
	// insert into taskCategorys table
	query := `INSERT INTO task_categories(name) VALUES($1) RETURNING id;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	err = db.Conn.QueryRow(taskCategory.Name).Scan(&id)
	if err != nil {
		return err
	}
	taskCategory.ID = id
	return nil
}

func (db Database) GetTaskCategoryByID(taskCategoryID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.TaskCategory, error) {
	taskCategory := models.TaskCategory{}
	query := `SELECT * FROM task_categories WHERE id = $1;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return taskCategory, err
	}

	defer stmt.Close()

	row := stmt.QueryRow(taskCategoryID)
	switch err := row.Scan(&taskCategory.ID, &taskCategory.Name); err {
	case sql.ErrNoRows:
		return taskCategory, ErrNoMatch
	default:
		return taskCategory, err
	}
}

func (db Database) DeleteTaskCategory(taskCategoryID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) error {
	isManager, err := db.IsManager(r, tokenAuth)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	query := `DELETE FROM task_categories WHERE id = $1`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(taskCategoryID)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

func (db Database) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.TaskCategory, error) {
	taskCategory := models.TaskCategory{}
	isManager, err := db.IsManager(r, tokenAuth)
	if err != nil {
		return taskCategory, err
	}
	if !isManager {
		return taskCategory, errors.New("you are not the manager")
	}
	query := `UPDATE task_categories SET name=$1 WHERE id=$2 RETURNING *;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return taskCategory, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(taskCategoryData.Name, taskCategoryData.ID).Scan(&taskCategory.ID, &taskCategory.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return taskCategory, ErrNoMatch
		}
		return taskCategory, err
	}
	return taskCategory, nil
}
