package controller

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
)

type TaskCategoryController struct {
	TaskCategoryRepository *repository.TaskCategoryRepository
}

func NewTaskCategoryController(taskCategoryRepository *repository.TaskCategoryRepository) *TaskCategoryController {
	return &TaskCategoryController{TaskCategoryRepository: taskCategoryRepository}
}

func (c *TaskCategoryController) GetAllTaskCategories(ctx context.Context) (models.TaskCategorySlice, error) {
	taskCategories, err := c.TaskCategoryRepository.GetAllTaskCategories(ctx)
	if err != nil {
		return taskCategories, err
	}
	return taskCategories, nil
}

func (c *TaskCategoryController) AddTaskCategory(taskCategory *models.TaskCategory, ctx context.Context) error {
	err := c.TaskCategoryRepository.AddTaskCategory(taskCategory, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskCategoryController) GetTaskCategoryByID(taskCategoryID int, ctx context.Context) (*models.TaskCategory, error) {
	taskCategory, err := c.TaskCategoryRepository.GetTaskCategoryByID(taskCategoryID, ctx)
	if err != nil {
		return taskCategory, err
	}
	return taskCategory, nil
}

func (c *TaskCategoryController) DeleteTaskCategory(taskCategoryID int, ctx context.Context) error {
	err := c.TaskCategoryRepository.DeleteTaskCategory(taskCategoryID, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskCategoryController) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, ctx context.Context) (*models.TaskCategory, error) {
	taskCategory, err := c.TaskCategoryRepository.GetTaskCategoryByID(taskCategoryID, ctx)
	if err != nil {
		return taskCategory, err
	}
	taskCategory.Name = taskCategoryData.Name

	taskCategoryUpdated, err := c.TaskCategoryRepository.UpdateTaskCategory(taskCategory, ctx)
	if err != nil {
		return taskCategoryUpdated, err
	}
	return taskCategoryUpdated, nil
}
