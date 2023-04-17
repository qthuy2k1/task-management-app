package repository

// import (
// 	"net/http"

// 	"github.com/go-chi/jwtauth/v5"
// 	"github.com/lestrrat-go/jwx/v2/jwt"
// 	"github.com/qthuy2k1/task-management-app/models"
// 	"github.com/stretchr/testify/mock"
// )

// type MockUserService struct {
// 	mock.Mock
// }

// func (m *MockUserService) GetAllUsers(r *http.Request, tokenAuth *jwtauth.JWTAuth) (*models.UserList, error) {
// 	args := m.Called(r, tokenAuth)
// 	return args.Get(0).(*models.UserList), args.Error(1)
// }

// func (m *MockUserService) AddUser(user *models.User) error {
// 	args := m.Called(user)
// 	return args.Error(0)
// }
// func (m *MockUserService) GetUserByID(userID int) (models.User, error) {
// 	args := m.Called(userID)
// 	var user models.User
// 	if args.Error(1) == nil {
// 		user = args.Get(0).(models.User)
// 	}
// 	return user, args.Error(1)
// }

// func (m *MockUserService) GetUserByEmail(userEmail string) (models.User, error) {
// 	args := m.Called(userEmail)
// 	return args.Get(0).(models.User), args.Error(1)
// }
// func (m *MockUserService) DeleteUser(userID int, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
// 	args := m.Called(userID, r, tokenAuth, token)
// 	return args.Error(1)
// }

// func (m *MockUserService) UpdateUser(userID int, userData models.User) (models.User, error) {
// 	args := m.Called(userID, userData)
// 	return args.Get(0).(models.User), args.Error(1)
// }

// func (m *MockUserService) UpdateRole(userID int, role string) (models.User, error) {
// 	args := m.Called(userID, role)
// 	return args.Get(0).(models.User), args.Error(1)
// }
// func (m *MockUserService) IsValidEmail(email string) bool {
// 	args := m.Called(email)
// 	return args.Bool(0)
// }

// func (m *MockUserService) IsValidPassword(password string) bool {
// 	args := m.Called(password)
// 	return args.Bool(0)
// }

// func (m *MockUserService) CompareEmailAndPassword(email, password string, r *http.Request, tokenAuth *jwtauth.JWTAuth) (bool, error) {
// 	args := m.Called(email, password, r, tokenAuth)
// 	return args.Bool(0), args.Error(1)
// }
