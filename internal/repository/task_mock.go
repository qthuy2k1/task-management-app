package repository

// import (
// 	"net/http"

// 	"github.com/go-chi/jwtauth/v5"
// 	"github.com/lestrrat-go/jwx/v2/jwt"
// 	"github.com/qthuy2k1/task-management-app/models"
// 	"github.com/stretchr/testify/mock"
// )

// type MockTaskService struct {
// 	mock.Mock
// }

// func (m *MockTaskService) GetAllTasks(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.TaskList, error) {
// 	args := m.Called(r, tokenAuth)
// 	return args.Get(0).(*models.TaskList), args.Error(1)
// }
// func (m *MockTaskService) AddTask(task *models.Task, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(task)
// 	return args.Error(0)
// }
// func (m *MockTaskService) GetTaskByID(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.Task, error) {
// 	args := m.Called(taskID, r, tokenAuth)
// 	return args.Get(0).(models.Task), args.Error(1)
// }
// func (m *MockTaskService) DeleteTask(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(taskID, r, tokenAuth, token)
// 	return args.Error(0)
// }
// func (m *MockTaskService) UpdateTask(taskID int, taskData models.Task, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (models.Task, error) {
// 	args := m.Called(taskID, taskData)
// 	return args.Get(0).(models.Task), args.Error(1)
// }

// func (m *MockTaskService) LockTask(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(taskID)
// 	return args.Error(0)
// }

// func (m *MockTaskService) UnLockTask(taskID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(taskID)
// 	return args.Error(0)
// }
// func (m *MockTaskService) ImportTaskDataFromCSV(path string) (models.TaskList, error) {
// 	args := m.Called(path)
// 	return args.Get(0).(models.TaskList), args.Error(1)
// }
