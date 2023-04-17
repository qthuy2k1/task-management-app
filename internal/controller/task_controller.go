package controller

// import (
// 	"context"
// 	"errors"
// 	"time"

// 	"github.com/qthuy2k1/task-management-app/internal/models/gen"
// 	"github.com/qthuy2k1/task-management-app/internal/repository"
// )

// type TaskController struct {
// 	TaskRepository *repository.TaskRepository
// }

// func NewTaskController(taskRepository *repository.TaskRepository) *TaskController {
// 	return &TaskController{TaskRepository: taskRepository}
// }

// func (c *TaskController) ValidateCreateTaskRequest(task *models.Task) error {
// 	// Validate that the task has a non-empty name
// 	if task.Name == "" {
// 		return errors.New("Task name is required")
// 	}
// 	return nil
// }

// func (c *TaskController) CreateTask(ctx context.Context, task *models.Task, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	userID, err := c.TaskRepository.GetUserIDFromToken(tokenAuth, token)
// 	if err != nil {
// 		return err
// 	}

// 	// Check if the user is a manager
// 	isManager, err := c.TaskRepository.IsManager(ctx, userID)
// 	if err != nil {
// 		return err
// 	}
// 	if !isManager {
// 		return errors.New("You are not authorized to create tasks")
// 	}

// 	// Set the task creation and update times
// 	now := time.Now()
// 	task.CreatedAt = now
// 	task.UpdatedAt = now

// 	// Insert the task into the database
// 	err = c.TaskRepository.CreateTask(ctx, task)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
