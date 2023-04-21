package mockControllers

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/stretchr/testify/mock"
)

type MockUserTaskDetailService struct {
	mock.Mock
}

// Add 1 user to 1 task
func (m *MockUserTaskDetailService) AddUserToTask(userID int, taskID int, ctx context.Context) error {
	args := m.Called(userID, taskID, ctx)
	return args.Error(0)
}

// Delete 1 user from task
func (m *MockUserTaskDetailService) DeleteUserFromTask(userID int, taskID int, ctx context.Context) error {
	args := m.Called(userID, taskID, ctx)
	return args.Error(0)
}

// Get all the users that are assigned to the task
func (m *MockUserTaskDetailService) GetAllUsersAssignedToTask(taskID int, ctx context.Context) ([]models.User, error) {
	args := m.Called(taskID, ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

// Get all the tasks that are assigned to the user
func (m *MockUserTaskDetailService) GetAllTaskAssignedToUser(userID int, ctx context.Context) ([]models.Task, error) {
	args := m.Called(userID, ctx)
	return args.Get(0).([]models.Task), args.Error(1)
}
