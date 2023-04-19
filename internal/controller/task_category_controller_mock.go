package controller

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/stretchr/testify/mock"
)

type MockTaskCategoryService struct {
	mock.Mock
}

func (m *MockTaskCategoryService) GetAllTaskCategories(ctx context.Context) (models.TaskCategorySlice, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.TaskCategorySlice), args.Error(1)
}

func (m *MockTaskCategoryService) AddTaskCategory(taskCategory *models.TaskCategory, ctx context.Context) error {
	args := m.Called(taskCategory, ctx)
	return args.Error(0)
}

func (m *MockTaskCategoryService) GetTaskCategoryByID(taskCategoryID int, ctx context.Context) (*models.TaskCategory, error) {
	args := m.Called(taskCategoryID, ctx)
	return args.Get(0).(*models.TaskCategory), args.Error(1)
}

func (m *MockTaskCategoryService) DeleteTaskCategory(taskCategoryID int, ctx context.Context) error {
	args := m.Called(taskCategoryID, ctx)
	return args.Error(0)
}

func (m *MockTaskCategoryService) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, ctx context.Context) (*models.TaskCategory, error) {
	args := m.Called(taskCategoryID, taskCategoryData, ctx)
	return args.Get(0).(*models.TaskCategory), args.Error(1)
}
