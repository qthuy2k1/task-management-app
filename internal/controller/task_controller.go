package controller

import (
	"context"
	"errors"
	"time"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
)

type TaskController struct {
	TaskRepository *repository.TaskRepository
}

func NewTaskController(taskRepository *repository.TaskRepository) *TaskController {
	return &TaskController{TaskRepository: taskRepository}
}

func (c *TaskController) GetAllTasks(ctx context.Context) (models.TaskSlice, error) {
	tasks, err := c.TaskRepository.GetAllTasks(ctx)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}

func (c *TaskController) AddTask(task *models.Task, ctx context.Context) error {
	err := c.TaskRepository.AddTask(task, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskController) GetTaskByID(taskID int, ctx context.Context) (*models.Task, error) {
	task, err := c.TaskRepository.GetTaskByID(taskID, ctx)
	if err != nil {
		return task, err
	}
	return task, nil
}

func (c *TaskController) DeleteTask(taskID int, ctx context.Context) error {
	err := c.TaskRepository.DeleteTask(taskID, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskController) UpdateTask(taskID int, taskData models.Task, ctx context.Context, isManager bool) (*models.Task, error) {
	task, err := c.TaskRepository.GetTaskByID(taskID, ctx)
	if err != nil {
		return task, err
	}

	task.Name = taskData.Name
	task.Description = taskData.Description
	task.StartDate = taskData.StartDate
	task.EndDate = taskData.EndDate
	task.Status = taskData.Status
	if task.Status.String == "Lock" && !isManager {
		return task, errors.New("you are not the manager, cannot lock this task")
	}
	task.AuthorID = taskData.AuthorID
	task.UpdatedAt = time.Now()
	task.TaskCategoryID = taskData.TaskCategoryID

	taskUpdated, err := c.TaskRepository.UpdateTask(task, ctx)
	if err != nil {
		return taskUpdated, err
	}
	return taskUpdated, nil
}

func (c *TaskController) LockTask(taskID int, ctx context.Context) error {
	err := c.TaskRepository.LockTask(taskID, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskController) UnLockTask(taskID int, ctx context.Context) error {
	err := c.TaskRepository.UnLockTask(taskID, ctx)
	if err != nil {
		return err
	}
	return nil
}
