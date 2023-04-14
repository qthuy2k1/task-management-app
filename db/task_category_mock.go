package db

// import (
// 	"net/http"

// 	"github.com/go-chi/jwtauth/v5"
// 	"github.com/lestrrat-go/jwx/v2/jwt"
// 	"github.com/qthuy2k1/task-management-app/models"
// 	"github.com/stretchr/testify/mock"
// )

// type MockTaskCategoryService struct {
// 	mock.Mock
// }

// func (m *MockTaskCategoryService) GetAllTaskCategories(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.TaskCategoryList, error) {
// 	args := m.Called(r, tokenAuth)
// 	return args.Get(0).(*models.TaskCategoryList), args.Error(1)
// }

// func (m *MockTaskCategoryService) AddTaskCategory(taskCategory *models.TaskCategory, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(taskCategory)
// 	return args.Error(0)
// }

// func (m *MockTaskCategoryService) GetTaskCategoryByID(taskCategoryID int, r *http.Request, tokenAuth *jwtauth.JWTAuth) (models.TaskCategory, error) {
// 	args := m.Called(taskCategoryID, r, tokenAuth)
// 	return args.Get(0).(models.TaskCategory), args.Error(1)
// }

// func (m *MockTaskCategoryService) DeleteTaskCategory(taskCategoryID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(taskCategoryID)
// 	return args.Error(0)
// }

// func (m *MockTaskCategoryService) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (models.TaskCategory, error) {
// 	args := m.Called(taskCategoryID, taskCategoryData)
// 	return args.Get(0).(models.TaskCategory), args.Error(1)
// }
