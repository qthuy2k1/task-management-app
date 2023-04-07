package db

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/lib/pq"
	"github.com/qthuy2k1/task-management-app/models"
)

// define an enum for Status of task
type TaskStatus string

const (
	NotStarted TaskStatus = "Not Started"
	InProgress TaskStatus = "In Progress"
	Complete   TaskStatus = "Complete"
	Lock       TaskStatus = "Lock"
)

func (db Database) GetAllTasks(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.TaskList, error) {
	list := &models.TaskList{}
	query := `SELECT * FROM tasks;`
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
		var task models.Task
		err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)
		if err != nil {
			return list, err
		}
		list.Tasks = append(list.Tasks, task)
	}
	return list, nil
}
func (db Database) AddTask(task *models.Task, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}

	var id int
	var createdAt time.Time

	// insert into tasks table and get the id, create_at
	query := `insert into tasks(name, description, start_date, end_date, status, author_id, task_category_id) values($1, $2, $3, $4, $5, $6, $7) returning id, created_at;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	quotedName := pq.QuoteLiteral(task.Name)
	quotedDescription := pq.QuoteLiteral(task.Description)
	quotedStatus := pq.QuoteLiteral(task.Status)

	err = stmt.QueryRow(quotedName, quotedDescription, task.StartDate, task.EndDate, quotedStatus, task.AuthorID, task.TaskCategoryID).Scan(&id, &createdAt)
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
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return task, err
	}

	defer stmt.Close()

	row := stmt.QueryRow(taskID)
	switch err := row.Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID); err {
	case sql.ErrNoRows:
		return task, ErrNoMatch
	default:
		return task, err
	}
}
func (db Database) DeleteTask(taskId int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}

	query := `DELETE FROM tasks WHERE id = $1;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(taskId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}
func (db Database) UpdateTask(taskID int, taskData models.Task, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (models.Task, error) {
	task := models.Task{}

	isManager, err := db.IsManager(r, tokenAuth, token)
	if err != nil {
		return task, err
	}
	query := ""

	if isManager {
		query = `UPDATE tasks SET name=$1, description=$2, start_date=$3, end_date=$4, status=$5, author_id=$6, updated_at=$7, task_category_id=$8 WHERE id=$9 RETURNING *;`
	} else {
		query = `UPDATE tasks SET name=$1, description=$2, start_date=$3, end_date=$4, status=$5, author_id=$6, updated_at=$7, task_category_id=$8 WHERE id=$9 AND status NOT IN ("Lock") RETURNING *;`
	}

	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return task, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(taskData.Name, taskData.Description, taskData.StartDate, taskData.EndDate, taskData.Status, taskData.AuthorID, time.Now(), taskData.TaskCategoryID, taskID).Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, ErrNoMatch
		}
		return task, err
	}
	return task, nil
}

func (db Database) LockTask(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	task := models.Task{}
	query := `UPDATE tasks SET status=$1 WHERE id=$2 RETURNING *;`

	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	err = stmt.QueryRow(Lock, taskID).Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNoMatch
		}
		return err
	}
	return nil
}

func (db Database) UnLockTask(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	task := models.Task{}
	query := `UPDATE tasks SET status=$1 WHERE id=$2 RETURNING *;`
	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	err = stmt.QueryRow(InProgress, taskID).Scan(&task.ID, &task.Name, &task.Description, &task.StartDate, &task.EndDate, &task.Status, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt, &task.TaskCategoryID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNoMatch
		}
		return err
	}
	return nil
}

func (db Database) GetTaskFromCSV(path string) (models.TaskList, error) {
	// Create a slice to store the task data
	taskList := models.TaskList{}

	// Open the CSV file
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error:", err)
		return taskList, err
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)
	// Trim leading spaces around fields
	reader.TrimLeadingSpace = true

	// Read all the records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return taskList, err
	}

	// Layout for time.Time field in struct
	const layoutDate = "2006-01-02T15:04:05Z"

	// Loop through the records and create a new Task struct for each record
	for i, record := range records {
		// Skip the first index, which is the header name
		if i == 0 {
			continue
		}
		data := strings.Split(record[0], ",")
		// remove unexpected characters
		re := regexp.MustCompile("[^a-zA-Z0-9 ]+")

		// regex for Name
		name := re.ReplaceAllString(data[0], "")

		// regex for description
		description := re.ReplaceAllString(data[1], "")

		// convert start date from string to time.Time
		startDate, err := time.Parse(layoutDate, data[2])
		if err != nil {
			return taskList, err
		}

		// convert end date from string to time.Time
		endDate, err := time.Parse(layoutDate, data[3])
		if err != nil {
			return taskList, err
		}

		// regex for task status
		status := re.ReplaceAllString(data[4], "")

		// convert authorID from string to int
		authorID, err := strconv.Atoi(data[5])
		if err != nil {
			return taskList, err
		}

		// convert createdAt from string to time.Time
		createdAt, err := time.Parse(layoutDate, data[6])
		if err != nil {
			return taskList, err
		}

		// convert updatedAt from string to time.Time
		updatedAt, err := time.Parse(layoutDate, data[7])
		if err != nil {
			return taskList, err
		}

		// convert taskCategoryID from string to int
		taskCategoryID, err := strconv.Atoi(data[8])
		if err != nil {
			return taskList, err
		}

		task := models.Task{
			Name:           name,
			Description:    description,
			StartDate:      startDate,
			EndDate:        endDate,
			Status:         status,
			AuthorID:       authorID,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			TaskCategoryID: taskCategoryID,
		}
		taskList.Tasks = append(taskList.Tasks, task)
	}

	return taskList, nil
}
