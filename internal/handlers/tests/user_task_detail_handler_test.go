package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/internal/handlers"
	mockController "github.com/qthuy2k1/task-management-app/internal/mocks/controllers"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repositories"
	"github.com/volatiletech/null/v8"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(handlers.Secret), nil)
}

func TestAddUserTaskDetailHandler(t *testing.T) {
	// Create a new instance of the mock user task detail service
	mockUserTaskDetailService := &mockController.MockUserTaskDetailService{}

	// Create a new router
	r := chi.NewRouter()

	// Set up the route for the createUserTaskDetail handler
	r.Post("/tasks/{taskID}/add-user", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			render.Render(w, r, handlers.ServerErrorRenderer(fmt.Errorf("failed to parse form data")))
		}
		userID, err := strconv.Atoi(r.PostForm.Get("id"))
		if err != nil {
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("invalid user id")))
			return
		}
		taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
		if err != nil {
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("invalid task id")))
			return
		}

		tokenStr := r.Context().Value("token").(string)
		if tokenStr == "" {
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if err != nil {
			render.Render(w, r, handlers.ErrorRenderer(err))
			return
		}
		if token == nil {
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		err = mockUserTaskDetailService.AddUserToTask(userID, taskID, context.Background())
		if err != nil {
			render.Render(w, r, handlers.ErrorRenderer(err))
			return
		}
	})

	// Define the test cases as a slice of structs
	testCases := []struct {
		name           string
		userID         int
		taskID         int
		isManager      bool
		token          string
		expectedStatus int
		expectedError  error
	}{
		{
			name:           "Success - User added to task",
			userID:         1,
			taskID:         1,
			isManager:      true,
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:           "Invalid user ID",
			userID:         0,
			taskID:         1,
			isManager:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  fmt.Errorf("invalid user id"),
		},
		// Add more test cases here as needed
	}

	for _, tc := range testCases {
		// Create a new request with the test JWT token
		_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
			"email":    "test@example.com",
			"password": "password",
		})

		// Set up the mock service to return the expected error
		mockUserTaskDetailService.On("AddUserToTask", tc.userID, tc.taskID, context.Background()).Return(tc.expectedError)

		// Create a new request with the test user ID and task ID
		formData := url.Values{}
		formData.Set("id", strconv.Itoa(tc.userID))
		req := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d/add-user", tc.taskID), strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))

		// Add the token string to the context
		ctx := context.WithValue(req.Context(), "token", tokenStr)
		req = req.WithContext(ctx)

		// Create a new recorder for the response
		recorder := httptest.NewRecorder()

		// Call the handler with the test request and recorder
		r.ServeHTTP(recorder, req)

		// Check the response status code
		if recorder.Result().StatusCode != tc.expectedStatus {
			t.Errorf("%s: expected status code %d but got %d", tc.name, tc.expectedStatus, recorder.Result().StatusCode)
		}

		// Check the mock service was called with the expected arguments
		mockUserTaskDetailService.AssertCalled(t, "AddUserToTask", tc.userID, tc.taskID, context.Background())

		// Reset the mock service
		mockUserTaskDetailService.AssertExpectations(t)
	}
}
func TestDeleteUserTaskDetailHandler(t *testing.T) {
	// Create a new instance of the mock user task detail service
	mockUserTaskDetailService := &mockController.MockUserTaskDetailService{}

	// Create a new router
	r := chi.NewRouter()

	// Set up the route for the DeleteUserTaskDetailHandler handler
	r.Post("/tasks/{taskID}/delete-user", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form data.", http.StatusInternalServerError)
			return
		}
		// Get the user ID from the form data
		userIDStr := r.PostForm.Get("id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Get the task ID from the URL params
		taskIDStr := chi.URLParam(r, "taskID")
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		// Delete the user from the task using the task detail service
		err = mockUserTaskDetailService.DeleteUserFromTask(userID, taskID, context.Background())
		if userID == 0 {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
		if taskID == 0 {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err != nil {
			if err == repositories.ErrNoMatch {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Return a success response
		w.WriteHeader(http.StatusOK)
	})

	// Define the test cases as a slice of structs
	testCases := []struct {
		name           string
		userID         int
		taskID         int
		isManager      bool
		token          string
		expectedStatus int
		expectedError  error
	}{
		{
			name:           "Success - User removed from task",
			userID:         1,
			taskID:         1,
			isManager:      true,
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:           "Invalid user ID",
			userID:         0,
			taskID:         1,
			isManager:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  fmt.Errorf("invalid user id"),
		},
		// Add more test cases here as needed
	}

	for _, tc := range testCases {
		// Create a new request with the test JWT token
		_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
			"email":    "test@example.com",
			"password": "password",
		})

		// Set up the mock service to return the expected error
		mockUserTaskDetailService.On("DeleteUserFromTask", tc.userID, tc.taskID, context.Background()).Return(tc.expectedError)

		reqBody := fmt.Sprintf("id=%d", tc.userID)
		// Create a new request with the test user ID and task ID
		req := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d/delete-user", tc.taskID), strings.NewReader(reqBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Add the token string to the context
		ctx := context.WithValue(req.Context(), "token", tokenStr)
		req = req.WithContext(ctx)

		// Create a new recorder for the response
		recorder := httptest.NewRecorder()

		// Call the handler with the test request and recorder
		r.ServeHTTP(recorder, req)

		// Check the response status code
		if recorder.Result().StatusCode != tc.expectedStatus {
			t.Errorf("%s: expected status code %d but got %d", tc.name, tc.expectedStatus, recorder.Result().StatusCode)
		}

		// Check if the mock service was called at all
		mockUserTaskDetailService.AssertCalled(t, "DeleteUserFromTask", tc.userID, tc.taskID, context.Background())

		// Reset the mock service
		mockUserTaskDetailService.AssertExpectations(t)
	}
}
func TestGetAllUserAssignedToTaskHandler(t *testing.T) {
	// Create a new instance of the mock user task detail service
	mockUserTaskDetailService := &mockController.MockUserTaskDetailService{}

	// Create a new router
	r := chi.NewRouter()

	// Set up the route for the GetAllUserAsignnedToTaskHandler handler
	r.Get("/tasks/{taskID}/get-users", func(w http.ResponseWriter, r *http.Request) {
		// Get the task ID from the URL params
		userIDStr := chi.URLParam(r, "taskID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		// Get all the users assigned to the task using the task detail service
		tasks, err := mockUserTaskDetailService.GetAllUsersAssignedToTask(userID, context.Background())
		if userID == 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Marshal the users into JSON and write the response
		jsonBytes, err := json.Marshal(tasks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	// Define the test cases as a slice of structs
	testCases := []struct {
		name           string
		taskID         int
		expectedStatus int
		expectedJSON   string
		expectedError  error
	}{
		{
			name:           "Success - Users assigned to task",
			taskID:         1,
			expectedStatus: http.StatusOK,
			expectedJSON:   `[{"id":1,"name":"Alice","email":"alice@example.com","password":"password","role":"user"},{"id":2,"name":"Bob","email":"bob@example.com","password":"password","role":"admin"}]`,
			expectedError:  nil,
		},
		{
			name:           "Invalid task ID",
			taskID:         0,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   "invalid task ID\n",
			expectedError:  fmt.Errorf("invalid task ID"),
		},
		// Add more test cases here as needed
	}
	for _, tc := range testCases {
		// Set up the mock service to return the expected users and error
		mockUserTaskDetailService.On("GetAllUsersAssignedToTask", tc.taskID, context.Background()).Return([]models.User{
			{ID: 1, Name: "Alice", Email: "alice@example.com", Password: "password", Role: "user"},
			{ID: 2, Name: "Bob", Email: "bob@example.com", Password: "password", Role: "admin"},
		}, tc.expectedError)

		// Create a new request with the test task ID
		req := httptest.NewRequest("GET", fmt.Sprintf("/tasks/%d/get-users", tc.taskID), nil)

		// Create a new recorder for the response
		recorder := httptest.NewRecorder()

		// Call the GetAllUserAsignnedToTaskHandler handler function
		r.ServeHTTP(recorder, req)

		// Check that the response status code matches the expected value
		if recorder.Code != tc.expectedStatus {
			t.Errorf("%s: expected status %d but got %d", tc.name, tc.expectedStatus, recorder.Code)
		}

		// Check that the response body matches the expected JSON string
		if recorder.Body.String() != tc.expectedJSON {
			t.Errorf("%s: expected JSON %s but got %s", tc.name, tc.expectedJSON, recorder.Body.String())
		}

		// Check that the mock service was called with the expected arguments
		mockUserTaskDetailService.AssertCalled(t, "GetAllUsersAssignedToTask", tc.taskID, context.Background())
	}
}
func TestGetAllTaskAssignedToUser(t *testing.T) {
	// Create a new instance of the mock user task detail service
	mockUserTaskDetailService := &mockController.MockUserTaskDetailService{}

	// Create a new router
	r := chi.NewRouter()

	// Set up the route for the GetAllUserAsignnedToTaskHandler handler
	r.Get("/users/{userID}/get-tasks", func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the URL params
		userIDStr := chi.URLParam(r, "userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Get all the tasks assigned to the user using the task detail service
		tasks, err := mockUserTaskDetailService.GetAllTaskAssignedToUser(userID, context.Background())
		if userID == 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Marshal the tasks into JSON and write the response
		jsonBytes, err := json.Marshal(tasks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)

	})

	// Define the test cases as a slice of structs
	testCases := []struct {
		name           string
		userID         int
		expectedStatus int
		expectedJSON   string
		expectedError  error
	}{
		{
			name:           "Success - Tasks assigned to user",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedJSON:   `[{"id":1,"name":"Task 1","description":"Description of task 1","start_date":"2022-12-01T12:00:00Z","end_date":"2022-12-02T12:00:00Z","status":"in progress","author_id":1,"created_at":"2022-12-01T12:00:00Z","updated_at":"2022-12-02T12:00:00Z","task_category_id":1},{"id":2,"name":"Task 2","description":"Description of task 2","start_date":"2022-12-03T12:00:00Z","end_date":"2022-12-04T12:00:00Z","status":"completed","author_id":1,"created_at":"2022-12-03T12:00:00Z","updated_at":"2022-12-04T12:00:00Z","task_category_id":1}]`,
			expectedError:  nil,
		},
		{
			name:           "Invalid user ID",
			userID:         0,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   "invalid user ID",
			expectedError:  fmt.Errorf("invalid user ID"),
		},
		// Add more test cases here as needed
	}
	for _, tc := range testCases {
		// Set up the mock service to return the expected users and error
		mockUserTaskDetailService.On("GetAllTaskAssignedToUser", tc.userID, context.Background()).Return([]models.Task{
			{ID: 1, Name: "Task 1", Description: "Description of task 1", StartDate: time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC), EndDate: time.Date(2022, 12, 2, 12, 0, 0, 0, time.UTC), Status: null.NewString("in progress", true), CreatedAt: time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2022, 12, 2, 12, 0, 0, 0, time.UTC), AuthorID: 1, TaskCategoryID: 1},
			{ID: 2, Name: "Task 2", Description: "Description of task 2", StartDate: time.Date(2022, 12, 3, 12, 0, 0, 0, time.UTC), EndDate: time.Date(2022, 12, 4, 12, 0, 0, 0, time.UTC), Status: null.NewString("completed", true), CreatedAt: time.Date(2022, 12, 3, 12, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2022, 12, 4, 12, 0, 0, 0, time.UTC), AuthorID: 1, TaskCategoryID: 1},
		}, tc.expectedError)

		// Create a new request with the test user ID
		req := httptest.NewRequest("GET", fmt.Sprintf("/users/%d/get-tasks", tc.userID), nil)

		// Create a new recorder for the response
		recorder := httptest.NewRecorder()

		// Call the GetAllTaskAssignedToUserHandler handler function
		r.ServeHTTP(recorder, req)

		// Check that the response status code matches the expected value
		if recorder.Code != tc.expectedStatus {
			t.Errorf("%s: expected status %d but got %d", tc.name, tc.expectedStatus, recorder.Code)
		}

		// Check that the response body matches the expected JSON string
		if strings.TrimRight(recorder.Body.String(), "\n\t\r") != tc.expectedJSON {
			t.Errorf("%s: expected JSON %s but got %s", tc.name, tc.expectedJSON, recorder.Body.String())
		}

		// Check that the mock service was called with the expected arguments
		mockUserTaskDetailService.AssertCalled(t, "GetAllTaskAssignedToUser", tc.userID, context.Background())
	}
}
