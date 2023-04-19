package controller

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) GetAllTasks(ctx context.Context) (models.TaskSlice, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.TaskSlice), args.Error(1)
}
func (m *MockTaskService) AddTask(task *models.Task, ctx context.Context) error {
	args := m.Called(task, ctx)
	return args.Error(0)
}
func (m *MockTaskService) GetTaskByID(taskID int, ctx context.Context) (*models.Task, error) {
	args := m.Called(taskID, ctx)
	return args.Get(0).(*models.Task), args.Error(1)
}
func (m *MockTaskService) DeleteTask(taskID int, ctx context.Context) error {
	args := m.Called(taskID, ctx)
	return args.Error(0)
}
func (m *MockTaskService) UpdateTask(taskID int, taskData models.Task, ctx context.Context) (models.Task, error) {
	args := m.Called(taskID, taskData, ctx)
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockTaskService) LockTask(taskID int, ctx context.Context) error {
	args := m.Called(taskID, ctx)
	return args.Error(0)
}

func (m *MockTaskService) UnLockTask(taskID int, ctx context.Context) error {
	args := m.Called(taskID, ctx)
	return args.Error(0)
}
func (m *MockTaskService) ImportTaskDataFromCSV(path string) ([]models.Task, error) {
	args := m.Called(path)
	return args.Get(0).([]models.Task), args.Error(1)
}
