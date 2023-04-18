package repository

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TaskRepository struct {
	Database *Database
}

func NewTaskRepository(database *Database) *TaskRepository {
	return &TaskRepository{Database: database}
}

// define an enum for Status of task
type TaskStatus string

const (
	NotStarted TaskStatus = "Not Started"
	InProgress TaskStatus = "In Progress"
	Complete   TaskStatus = "Complete"
	Lock       TaskStatus = "Lock"
)

func (re *TaskRepository) GetAllTasks(ctx context.Context) (models.TaskSlice, error) {
	tasks, err := models.Tasks().All(ctx, re.Database.Conn)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}

// Adds a new task to the database
func (re *TaskRepository) AddTask(task *models.Task, ctx context.Context) error {
	err := task.Insert(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// Retrieves a task from the database by ID
func (re *TaskRepository) GetTaskByID(taskID int, ctx context.Context) (*models.Task, error) {
	task, err := models.Tasks(Where("id = ?", taskID)).One(ctx, re.Database.Conn)

	if err != nil {
		return task, err
	}
	return task, nil
}

// Deletes a task from the database by ID
func (re *TaskRepository) DeleteTask(taskID int, ctx context.Context) error {
	rowsAff, err := models.Tasks(Where("id = ?", taskID)).DeleteAll(ctx, re.Database.Conn)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoMatch
	}
	return err
}

// Updates a task in the database by ID
func (re *TaskRepository) UpdateTask(task *models.Task, ctx context.Context) (*models.Task, error) {
	rowsAff, err := task.Update(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return task, err
	}
	if rowsAff == 0 {
		return task, ErrNoMatch
	}
	return task, nil

}

// Locks a task in the database by ID
func (re *TaskRepository) LockTask(taskID int, ctx context.Context) error {
	rowsAff, err := models.Tasks(Where("id = ?", taskID)).UpdateAll(ctx, re.Database.Conn, models.M{"status": Lock})
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoMatch
	}
	return nil
}

// Unlocks a task in the database by ID
func (re *TaskRepository) UnLockTask(taskID int, ctx context.Context) error {
	rowsAff, err := models.Tasks(Where("id = ?", taskID)).UpdateAll(ctx, re.Database.Conn, models.M{"status": InProgress})
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
