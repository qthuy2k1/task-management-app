package mockControllers

import (
	"context"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetAllUsers(ctx context.Context) (models.UserSlice, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.UserSlice), args.Error(1)
}

func (m *MockUserService) AddUser(user *models.User, ctx context.Context) error {
	args := m.Called(user, ctx)
	return args.Error(0)
}
func (m *MockUserService) GetUserByID(userID int, ctx context.Context) (*models.User, error) {
	args := m.Called(userID, ctx)
	var user *models.User
	if args.Error(1) == nil {
		user = args.Get(0).(*models.User)
	}
	return user, args.Error(1)
}

func (m *MockUserService) GetUserByEmail(userEmail string, ctx context.Context) (*models.User, error) {
	args := m.Called(userEmail)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserService) DeleteUser(userID int, ctx context.Context) (int64, error) {
	args := m.Called(userID, ctx)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockUserService) UpdateUser(userID int, userData models.User, ctx context.Context) (*models.User, error) {
	args := m.Called(userID, userData, ctx)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateRole(userID int, role string, ctx context.Context) (*models.User, error) {
	args := m.Called(userID, role, ctx)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUsersManager(ctx context.Context) (models.UserSlice, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.UserSlice), args.Error(1)
}
