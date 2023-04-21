package repositories

import (
	"context"
	"database/sql"
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

func (re *TaskRepository) GetAllTasks(ctx context.Context, pageNumber int, pageSize int, sortField string, sortOrder string) (models.TaskSlice, error) {
	fmt.Println(sortField)
	queryMods := []QueryMod{
		OrderBy(fmt.Sprintf(sortField + " " + sortOrder)),
		Offset((pageNumber - 1) * pageSize),
		Limit(pageSize),
	}
	// rows, err := models.Tasks(OrderBy("? ?", sortField, sortOrder), Limit(pageSize), Offset((pageNumber-1)*pageSize)).All(ctx, re.Database.Conn)
	rows, err := models.Tasks(queryMods...).All(ctx, re.Database.Conn)
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

	// if err := rows.Err(); err != nil {
	//     return nil, err
	// }

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
