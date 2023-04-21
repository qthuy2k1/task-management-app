package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (re *TaskRepository) GetAllTasks(ctx context.Context, filterValues map[string]interface{}) (models.TaskSlice, error) {
	var (
		sortField string
		sortOrder string
	)
	query := []QueryMod{Where("1=1")}
	for field, value := range filterValues {
		switch field {
		case "id":
			valueConv, ok := value.(int)
			if !ok {
				return nil, errors.New("cannot convert interface{} id to int")
			}
			query = append(query, Where("id = ?", valueConv))
		case "name":
			valueConv, ok := value.(string)
			if !ok {
				return nil, errors.New("cannot convert interface{} name to string")
			}
			query = append(query, Where("name LIKE ?", "%"+valueConv+"%"))
		case "description":
			valueConv, ok := value.(string)
			if !ok {
				return nil, errors.New("cannot convert interface{} description to string")
			}
			query = append(query, Where("description LIKE ?", "%"+valueConv+"%"))
		case "status":
			valueConv, ok := value.(string)
			if !ok {
				return nil, errors.New("cannot convert interface{} status to string")
			}
			query = append(query, Where("status LIKE ?", "%"+valueConv+"%"))
		case "author_id":
			valueConv, ok := value.(int)
			if !ok {
				return nil, errors.New("cannot convert string author_id to int")
			}
			query = append(query, Where("author_id = ?", valueConv))
		case "task_category_id":
			valueConv, ok := value.(int)
			if !ok {
				return nil, errors.New("cannot convert string task_category_id to int")
			}
			query = append(query, Where("task_category_id = ?", valueConv))
		case "field":
			sortField, ok := value.(string)
			if !ok {
				return nil, errors.New("cannot convert interface{} sortfield to string")
			}
		case "order":
			sortOrder, ok := value.(string)
			if !ok {
				return nil, errors.New("cannot convert interface{} sortorder to string")
			}
			// query = append(query, OrderBy(fmt.Sprintf(sortField+" "+sortOrder)))
		case "page":
			pageNumber, ok := value.(int)
			if !ok {
				return nil, errors.New("cannot convert interface{} page to int")
			}
			query = append(query, Offset(pageNumber))
		case "size":
			pageSize, ok := value.(int)
			if !ok {
				return nil, errors.New("cannot convert interface{} size to int")
			}
			query = append(query, Limit(pageSize))
		}
	}
	query = append(query, OrderBy(fmt.Sprintf(sortField+" "+sortOrder)))
	rows, err := models.Tasks(query...).All(ctx, re.Database.Conn)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoMatch
		}
		return nil, err
	}

	tasks := make(models.TaskSlice, 0)
	for _, x := range rows {
		// var task *models.Task
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, x)
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
		if err == sql.ErrNoRows {
			return task, ErrNoMatch
		}
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

// Get the task category of a task
func (re *TaskRepository) GetTaskCategoryOfTask(taskID int, ctx context.Context) (*models.TaskCategory, error) {
	task, err := models.Tasks(
		Where("id = ?", taskID),
	).One(ctx, re.Database.Conn)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoMatch
		}
		return nil, err
	}

	taskCategory, err := task.TaskCategory().One(ctx, re.Database.Conn)
	if err != nil {
		if err == sql.ErrNoRows {
			return taskCategory, ErrNoMatch
		}
		return taskCategory, err
	}
	return taskCategory, nil
}

// Get tasks filter by name in query
func (re *TaskRepository) GetTasksByName(name string, ctx context.Context) (models.TaskSlice, error) {
	tasks, err := models.Tasks(Where("name LIKE ?", "%"+name+"%")).All(ctx, re.Database.Conn)
	if err != nil {
		if err == sql.ErrNoRows {
			return tasks, ErrNoMatch
		}
		return tasks, err
	}
	return tasks, nil
}

// Total number of tasks
func (re *TaskRepository) GetTaskCount(ctx context.Context) (int64, error) {
	count, err := models.Tasks().Count(ctx, re.Database.Conn)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// Total number of filtered status tasks
func (re *TaskRepository) CountFilteredStatusTask(status string, ctx context.Context) (int64, error) {
	count, err := models.Tasks(Where("status = ?", status)).Count(ctx, re.Database.Conn)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// func (re *TaskRepository) FilterTasks(, ctx context.Context) (models.TaskSlice, error) {
// 	// Build a query using SQLBoiler's query builder.

// 	// Execute the query and return the results.
// 	tasks, err := models.Tasks(query...).All(ctx, re.Database.Conn)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return tasks, nil
// }
