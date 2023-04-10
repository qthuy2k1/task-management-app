package db

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/lib/pq"
	"github.com/qthuy2k1/task-management-app/models"
)

// Gets all task categories from the database
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

// Adds a new task category to the database
func (db Database) AddTaskCategory(taskCategory *models.TaskCategory, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	if taskCategory.Name == "" {
		return errors.New("bad request")
	}
	isManager, err := db.IsManager(r, tokenAuth, token)
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

	quotedName := pq.QuoteLiteral(taskCategory.Name)
	err = stmt.QueryRow(quotedName).Scan(&id)
	if err != nil {
		return err
	}
	taskCategory.ID = id
	return nil
}

// Gets a task category from the database by ID
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

// Deletes a task category from the database by ID
func (db Database) DeleteTaskCategory(taskCategoryID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(r, tokenAuth, token)
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

// Updates a task category in the database by ID
func (db Database) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (models.TaskCategory, error) {
	taskCategory := models.TaskCategory{}
	isManager, err := db.IsManager(r, tokenAuth, token)
	if err != nil {
		return taskCategory, err
	}
	if !isManager {
		return taskCategory, errors.New("you are not the manager")
	}
	taskCategory, err = db.GetTaskCategoryByID(taskCategoryID, r, tokenAuth)
	if err != nil {
		if err == sql.ErrNoRows {
			return taskCategory, ErrNoMatch
		}
		return taskCategory, err
	}
	query := `UPDATE task_categories SET name=$1 WHERE id=$2 RETURNING *;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return taskCategory, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(taskCategoryData.Name, taskCategory.ID).Scan(&taskCategory.ID, &taskCategory.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return taskCategory, ErrNoMatch
		}
		return taskCategory, err
	}
	return taskCategory, nil
}

// Import task categories data from a CSV file
func (db Database) ImportTaskCategoryDataFromCSV(path string) (models.TaskCategoryList, error) {
	// Create a slice to store the taskCategory data
	taskCategoryList := models.TaskCategoryList{}

	// Open the CSV file
	file, err := os.Open("./data/task-category.csv")
	if err != nil {
		return taskCategoryList, err
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true // Trim leading spaces around fields

	// Read all the records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return taskCategoryList, err
	}

	// Loop through the records and create a new TaskCategory struct for each record
	for i, record := range records {
		// Skip the first index, which is the header name
		if i == 0 {
			continue
		}

		// remove unexpected characters
		re := regexp.MustCompile("[^a-zA-Z0-9]+")
		name := re.ReplaceAllString(record[0], "")

		taskCategory := models.TaskCategory{
			Name: name,
		}
		taskCategoryList.TaskCategories = append(taskCategoryList.TaskCategories, taskCategory)
	}

	return taskCategoryList, nil
}
