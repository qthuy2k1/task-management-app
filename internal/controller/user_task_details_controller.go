package controller

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
)

type UserTaskDetailController struct {
	UserTaskDetailRepository *repository.UserTaskDetailRepository
}

func NewUserTaskDetailController(userTaskDetailRepository *repository.UserTaskDetailRepository) *UserTaskDetailController {
	return &UserTaskDetailController{UserTaskDetailRepository: userTaskDetailRepository}
}

func (c *UserTaskDetailController) AddUserToTask(userID, taskID int, ctx context.Context) error {
	err := c.UserTaskDetailRepository.AddUserToTask(userID, taskID)
	if err != nil {
		return err
	}
	return nil
}

func (c *UserTaskDetailController) DeleteUserFromTask(userID, taskID int, ctx context.Context) error {
	err := c.UserTaskDetailRepository.DeleteUserFromTask(userID, taskID)
	if err != nil {
		return err
	}
	return nil
}

func (c *UserTaskDetailController) GetAllUsersAssignedToTask(taskID int) ([]models.User, error) {
	users, err := c.UserTaskDetailRepository.GetAllUsersAssignedToTask(taskID)
	if err != nil {
		return users, err
	}
	return users, nil
}

func (c *UserTaskDetailController) GetAllTaskAssignedToUser(userID int) ([]models.Task, error) {
	tasks, err := c.UserTaskDetailRepository.GetAllTaskAssignedToUser(userID)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}
