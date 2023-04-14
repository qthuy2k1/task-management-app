package db

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/qthuy2k1/task-management-app/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// define an enum for Status of task
type TaskStatus string

const (
	NotStarted TaskStatus = "Not Started"
	InProgress TaskStatus = "In Progress"
	Complete   TaskStatus = "Complete"
	Lock       TaskStatus = "Lock"
)

func (db Database) GetAllTasks(ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.TaskSlice, error) {
	tasks, err := models.Tasks().All(ctx, db.Conn)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}

// Adds a new task to the database
func (db Database) AddTask(task *models.Task, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}

	var createdAt time.Time
	task.CreatedAt = createdAt
	task.UpdatedAt = createdAt

	err = task.Insert(ctx, db.Conn, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// Retrieves a task from the database by ID
func (db Database) GetTaskByID(taskID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.Task, error) {
	task, err := models.Tasks(Where("id = ?", taskID)).One(ctx, db.Conn)

	if err != nil {
		return task, err
	}
	return task, nil
}

// Deletes a task from the database by ID
func (db Database) DeleteTask(taskID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}

	rowsAff, err := models.Tasks(Where("id = ?", taskID)).DeleteAll(ctx, db.Conn)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoMatch
	}
	return err
}

// Updates a task in the database by ID
func (db Database) UpdateTask(taskID int, taskData models.Task, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (models.Task, error) {
	task := models.Task{}

	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return task, err
	}

	if !isManager && taskData.Status.String == "Lock" {
		return task, errors.New("you are not manager, cannot lock this task.")
	}
	task.Name = taskData.Name
	task.Description = taskData.Description
	task.StartDate = taskData.StartDate
	task.EndDate = taskData.EndDate
	task.Status = taskData.Status
	task.AuthorID = taskData.AuthorID
	task.UpdatedAt = time.Now()
	task.TaskCategoryID = taskData.TaskCategoryID

	rowsAff, err := task.Update(ctx, db.Conn, boil.Infer())
	if err != nil {
		return task, err
	}
	if rowsAff == 0 {
		return task, ErrNoMatch
	}
	return task, nil

}

// Locks a task in the database by ID
func (db Database) LockTask(taskID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	rowsAff, err := models.Tasks(Where("id = ?", taskID)).UpdateAll(ctx, db.Conn, models.M{"status": Lock})
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoMatch
	}
	return nil
}

// Unlocks a task in the database by ID
func (db Database) UnLockTask(taskID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	rowsAff, err := models.Tasks(Where("id = ?", taskID)).UpdateAll(ctx, db.Conn, models.M{"status": InProgress})
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoMatch
	}
	return nil
}

// Import tasks data from a CSV file
// func (db Database) ImportTaskDataFromCSV(path string) (models.TaskSlice, error) {
// 	// Create a slice to store the task data
// 	taskList := models.TaskList{}

// 	// Open the CSV file
// 	file, err := os.Open(path)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return taskList, err
// 	}
// 	defer file.Close()

// 	// Create a new CSV reader
// 	reader := csv.NewReader(file)
// 	// Trim leading spaces around fields
// 	reader.TrimLeadingSpace = true

// 	// Read all the records from the CSV file
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return taskList, err
// 	}

// 	// Layout for time.Time field in struct
// 	const layoutDate = "2006-01-02T15:04:05Z"

// 	// Loop through the records and create a new Task struct for each record
// 	for i, record := range records {
// 		// Skip the first index, which is the header name
// 		if i == 0 {
// 			continue
// 		}
// 		data := strings.Split(record[0], ",")
// 		// remove unexpected characters
// 		re := regexp.MustCompile("[^a-zA-Z0-9 ]+")

// 		// regex for Name
// 		name := re.ReplaceAllString(data[0], "")

// 		// regex for description
// 		description := re.ReplaceAllString(data[1], "")

// 		// convert start date from string to time.Time
// 		startDate, err := time.Parse(layoutDate, data[2])
// 		if err != nil {
// 			return taskList, err
// 		}

// 		// convert end date from string to time.Time
// 		endDate, err := time.Parse(layoutDate, data[3])
// 		if err != nil {
// 			return taskList, err
// 		}

// 		// regex for task status
// 		status := re.ReplaceAllString(data[4], "")

// 		// convert authorID from string to int
// 		authorID, err := strconv.Atoi(data[5])
// 		if err != nil {
// 			return taskList, err
// 		}

// 		// convert createdAt from string to time.Time
// 		createdAt, err := time.Parse(layoutDate, data[6])
// 		if err != nil {
// 			return taskList, err
// 		}

// 		// convert updatedAt from string to time.Time
// 		updatedAt, err := time.Parse(layoutDate, data[7])
// 		if err != nil {
// 			return taskList, err
// 		}

// 		// convert taskCategoryID from string to int
// 		taskCategoryID, err := strconv.Atoi(data[8])
// 		if err != nil {
// 			return taskList, err
// 		}

// 		task := models.Task{
// 			Name:           name,
// 			Description:    description,
// 			StartDate:      startDate,
// 			EndDate:        endDate,
// 			Status:         status,
// 			AuthorID:       authorID,
// 			CreatedAt:      createdAt,
// 			UpdatedAt:      updatedAt,
// 			TaskCategoryID: taskCategoryID,
// 		}
// 		taskList.Tasks = append(taskList.Tasks, task)
// 	}

// 	return taskList, nil
// }
