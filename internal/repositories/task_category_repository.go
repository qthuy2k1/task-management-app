package repositories

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TaskCategoryRepository struct {
	Database *Database
}

func NewTaskCategoryRepository(database *Database) *TaskCategoryRepository {
	return &TaskCategoryRepository{Database: database}
}

// Gets all task categories from the database
func (re *TaskCategoryRepository) GetAllTaskCategories(ctx context.Context) (models.TaskCategorySlice, error) {
	taskCategories, err := models.TaskCategories().All(ctx, re.Database.Conn)
	if err != nil {
		return taskCategories, err
	}
	return taskCategories, nil
}

// Adds a new task category to the database
func (re *TaskCategoryRepository) AddTaskCategory(taskCategory *models.TaskCategory, ctx context.Context) error {
	err := taskCategory.Insert(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// Gets a task category from the database by ID
func (re *TaskCategoryRepository) GetTaskCategoryByID(taskCategoryID int, ctx context.Context) (*models.TaskCategory, error) {
	taskCategory, err := models.TaskCategories(Where("id = ?", taskCategoryID)).One(ctx, re.Database.Conn)
	if err != nil {
		return taskCategory, err
	}
	return taskCategory, nil
}

// Deletes a task category from the database by ID
func (re *TaskCategoryRepository) DeleteTaskCategory(taskCategoryID int, ctx context.Context) error {
	rowsAff, err := models.TaskCategories(Where("id = ?", taskCategoryID)).DeleteAll(ctx, re.Database.Conn)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoMatch
	}
	return err
}

// Updates a task category in the database by ID
func (re *TaskCategoryRepository) UpdateTaskCategory(taskCategory *models.TaskCategory, ctx context.Context) (*models.TaskCategory, error) {
	rowsAff, err := taskCategory.Update(ctx, re.Database.Conn, boil.Infer())
	if err != nil {
		return taskCategory, err
	}
	if rowsAff == 0 {
		return taskCategory, ErrNoMatch
	}
	return taskCategory, nil
}
