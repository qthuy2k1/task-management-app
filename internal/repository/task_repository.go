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
