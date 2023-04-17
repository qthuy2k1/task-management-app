package repository

// import (
// 	"net/http"

// 	"github.com/go-chi/jwtauth/v5"
// 	"github.com/lestrrat-go/jwx/v2/jwt"
// 	"github.com/qthuy2k1/task-management-app/models"
// 	"github.com/stretchr/testify/mock"
// )

// type MockUserTaskDetailService struct {
// 	mock.Mock
// }

// // Add 1 user to 1 task
// func (m *MockUserTaskDetailService) AddUserToTask(userID int, taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(userID, taskID)
// 	return args.Error(0)
// }

// func (m *MockUserTaskDetailService) DeleteUserFromTask(userID int, taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(userID, taskID)
// 	return args.Error(0)
// }

// // Get all the users that are assigned to the task
// func (m *MockUserTaskDetailService) GetAllUsersAssignedToTask(taskID int) (*models.UserList, error) {
// 	args := m.Called(taskID)
// 	return args.Get(0).(*models.UserList), args.Error(1)
// }

// // Get all the tasks that are assigned to the user
// func (m *MockUserTaskDetailService) GetAllTaskAssignedToUser(userID int) (*models.TaskList, error) {
// 	args := m.Called(userID)
// 	return args.Get(0).(*models.TaskList), args.Error(1)
// }
