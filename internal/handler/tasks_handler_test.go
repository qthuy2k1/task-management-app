package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/internal/controller"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
	"github.com/stretchr/testify/mock"
	"github.com/volatiletech/null/v8"
)

func TestGetAllTasks(t *testing.T) {
	testCases := []struct {
		name           string
		pageNumber     int
		pageSize       int
		sortField      string
		sortOrder      string
		expectedStatus int
		expectedBody   string
		mockResults    models.TaskSlice
		mockError      error
	}{
		{
			name:           "Returns list of tasks",
			pageNumber:     2,
			pageSize:       1,
			sortField:      "id",
			sortOrder:      "asc",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"name":"Task 1","description":"Description of Task 1","start_date":"2023-04-20T13:00:00Z","end_date":"2023-04-21T13:00:00Z","status":"In Progress","author_id":1,"created_at":"2023-04-20T13:00:00Z","updated_at":"2023-04-20T13:00:00Z","task_category_id":1}]`,
			mockResults: models.TaskSlice{
				{
					ID:             1,
					Name:           "Task 1",
					Description:    "Description of Task 1",
					StartDate:      time.Date(2023, 4, 20, 13, 0, 0, 0, time.UTC),
					EndDate:        time.Date(2023, 4, 21, 13, 0, 0, 0, time.UTC),
					Status:         null.NewString("In Progress", true),
					AuthorID:       1,
					CreatedAt:      time.Date(2023, 4, 20, 13, 0, 0, 0, time.UTC),
					UpdatedAt:      time.Date(2023, 4, 20, 13, 0, 0, 0, time.UTC),
					TaskCategoryID: 1,
				},
			},
			mockError: nil,
		},
		{
			name:           "Returns error when GetAllTasks fails",
			pageNumber:     1,
			pageSize:       2,
			sortField:      "name",
			sortOrder:      "desc",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "mock error",
			mockResults:    nil,
			mockError:      errors.New("mock error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a new mock task service
			taskServiceMock := &controller.MockTaskService{}

			// Set up the mock task service to return a list of tasks
			taskServiceMock.On("GetAllTasks", mock.Anything, testCase.pageNumber, testCase.pageSize, testCase.sortField, testCase.sortOrder).Return(testCase.mockResults, testCase.mockError)

			// Create a new test request with pagination parameters
			req, err := http.NewRequest("GET", fmt.Sprintf("/tasks?page=%d&size=%d&sortfield=%s&sortorder=%s", testCase.pageNumber, testCase.pageSize, testCase.sortField, testCase.sortOrder), nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a new test response recorder
			rr := httptest.NewRecorder()

			// Call the getAllTasks function with the mock task service and the test request
			r := chi.NewRouter()
			r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
				pageNumber := 1
				pageSize := 2
				sortField := "id"
				sortOrder := "asc"

				// Retrieve the "page" query parameter from the request, if present
				if pageStr := r.URL.Query().Get("page"); pageStr != "" {
					pageNumber, _ = strconv.Atoi(pageStr)
				}

				// Retrieve the "size" query parameter from the request, if present
				if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
					pageSize, _ = strconv.Atoi(sizeStr)
				}

				// Retrieve the "sort" query parameter from the request, if present
				if sortFieldStr := r.URL.Query().Get("sortfield"); sortFieldStr != "" {
					sortField = sortFieldStr
				}
				if sortOrderStr := r.URL.Query().Get("sortorder"); sortOrderStr != "" {
					sortOrder = sortOrderStr
				}

				tasks, err := taskServiceMock.GetAllTasks(ctx, pageNumber, pageSize, sortField, sortOrder)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				jsonBytes, err := json.Marshal(tasks)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			})

			r.ServeHTTP(rr, req)

			// Assert that the response status code is as expected
			if status := rr.Code; status != testCase.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, testCase.expectedStatus)
			}

			// Assert that the response body matches the expected task list or error message
			response := strings.TrimRight(rr.Body.String(), "\n\t\r")
			if response != testCase.expectedBody {
				t.Errorf("Handler returned unexpected body: got %s want %s", response, testCase.expectedBody)
			}

			// Assert that the GetAllTasks method was called on the mock task service with the correct arguments
			taskServiceMock.AssertExpectations(t)
		})
	}
}

func TestGetTaskByID(t *testing.T) {
	// Define the mock task service
	taskServiceMock := &controller.MockTaskService{}

	// Define the test cases
	testCases := []struct {
		name           string
		taskID         int
		mockTask       *models.Task
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			taskID:         1,
			mockTask:       &models.Task{ID: 1, Name: "Task 1", Description: "Description of Task 1", StartDate: time.Now(), EndDate: time.Now().Add(time.Hour), Status: null.NewString("In Progress", true), AuthorID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(), TaskCategoryID: 1},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"Task 1","description":"Description of Task 1","start_date":"` + time.Now().Format(time.RFC3339Nano) + `","end_date":"` + time.Now().Add(time.Hour).Format(time.RFC3339Nano) + `","status":"In Progress","author_id":1,"created_at":"` + time.Now().Format(time.RFC3339Nano) + `","updated_at":"` + time.Now().Format(time.RFC3339Nano) + `","task_category_id":1}`,
		},
		{
			name:           "Task Not Found",
			taskID:         2,
			mockTask:       &models.Task{},
			mockErr:        repository.ErrNoMatch,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"status_text":"","message":"Resource not found"}`,
		},
		{
			name:           "Invalid Task ID",
			taskID:         0,
			mockTask:       &models.Task{},
			mockErr:        fmt.Errorf("invalid task ID"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status_text":"Bad request","message":"invalid task ID"}`,
		},
	}

	// Loop through the test cases and run each one
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the taskServiceMock to return the mock task and error for this test case
			taskServiceMock.On("GetTaskByID", tt.taskID, context.Background()).Return(tt.mockTask, tt.mockErr)

			// Create a new test request for GET /tasks/{id}
			req, err := http.NewRequest("GET", fmt.Sprintf("/tasks/%d", tt.taskID), nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a new test response recorder
			rr := httptest.NewRecorder()

			// Call the test request handler with the mock task service and the test request
			router := chi.NewRouter()
			router.Get("/tasks/{taskID}", func(w http.ResponseWriter, r *http.Request) {
				// Get the task ID from the URL parameter
				taskID := chi.URLParam(r, "taskID")

				// Parse the task ID as an integer
				id, err := strconv.Atoi(taskID)
				if err != nil {
					render.Render(w, r, ErrorRenderer(err))
					return
				}

				// Call the task service to get the task with the specified ID
				task, err := taskServiceMock.GetTaskByID(id, ctx)
				if err != nil {
					if err == repository.ErrNoMatch {
						render.Render(w, r, ErrNotFound)
					} else {
						render.Render(w, r, ErrorRenderer(err))
					}
					return
				}

				// Render the task as JSON and send it as the response body
				jsonBytes, err := json.Marshal(task)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			})
			router.ServeHTTP(rr, req)

			// Check that the response status code matches the expected status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check that the response body matches the expected body
			response := strings.TrimRight(rr.Body.String(), "\n\t\r")
			if response != tt.expectedBody {
				t.Errorf("Handler returned unexpected body: got %q want %q", response, tt.expectedBody)
			}
		})
	}
}
func TestAddTask(t *testing.T) {
	// Create a new mock task service
	taskServiceMock := &controller.MockTaskService{}

	// Create a new test router
	router := chi.NewRouter()

	// Register the createTask function as a handler for POST requests to /tasks
	router.Post("/tasks", func(w http.ResponseWriter, r *http.Request) {
		// Parse the task data from the request body
		var taskData models.Task
		err := json.NewDecoder(r.Body).Decode(&taskData)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		tokenStr := strings.Split(r.Header.Get("Authorization"), " ")[1]
		if tokenStr == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if err != nil {
			fmt.Println(err)
			render.Render(w, r, ErrorRenderer(err))
			return
		}
		if token == nil {
			fmt.Println(err)
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		// Call the task service to add the new task
		err = taskServiceMock.AddTask(&taskData, context.Background())
		if err != nil {
			render.Render(w, r, ErrorRenderer(err))
			return
		}

		// Render the new task as JSON and send it as the response body
		jsonBytes, err := json.Marshal(taskData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	// Define the mock task to use in the test
	mockTask := &models.Task{
		Name:           "Test Task",
		Description:    "This is a test task",
		StartDate:      time.Now().Local(),
		EndDate:        time.Now().Add(time.Hour).Local(),
		Status:         null.NewString("Working", true),
		AuthorID:       1,
		TaskCategoryID: 1,
	}
	_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
		"email":    "test@example.com",
		"password": "password",
	})
	// Set up the taskServiceMock to return nil when AddTask is called with the mock task
	taskServiceMock.On("AddTask", mockTask, context.Background()).Return(nil)

	// Encode the mock task as JSON and create a new test request for POST /tasks
	taskJSON, _ := json.Marshal(mockTask)
	req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(taskJSON))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the Content-Type header to "application/json"
	req.Header.Set("Content-Type", "application/json")

	// Create a new test response recorder
	rr := httptest.NewRecorder()

	// Call the test request handler with the mock task service and the test request
	router.ServeHTTP(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response body matches the expected JSON-encoded mock task
	response := strings.TrimRight(rr.Body.String(), "\n\t\r")
	if response != string(taskJSON) {
		t.Errorf("Handler returned unexpected body: got %q want %q", response, string(taskJSON))
	}
}
func TestDeleteTask(t *testing.T) {
	// Create a new mock task service
	taskServiceMock := &controller.MockTaskService{}

	// Create a new test router
	router := chi.NewRouter()

	// Register the deleteTask function as a handler for DELETE requests to /tasks/{taskId}
	router.Delete("/tasks/{taskID}", func(w http.ResponseWriter, r *http.Request) {
		// Get the task ID from the URL parameters
		taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
		if err != nil {
			render.Render(w, r, ErrBadRequest)
			return
		}

		tokenStr := strings.Split(r.Header.Get("Authorization"), " ")[1]
		if tokenStr == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if err != nil {
			fmt.Println(err)
			render.Render(w, r, ErrorRenderer(err))
			return
		}
		if token == nil {
			fmt.Println(err)
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}

		// Call the task service to delete the task
		err = taskServiceMock.DeleteTask(taskID, context.Background())
		if err != nil {
			if err == repository.ErrNoMatch {
				render.Render(w, r, ErrNotFound)
			} else {
				render.Render(w, r, ServerErrorRenderer(err))
			}
			return
		}

	})
	_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
		"email":    "test@example.com",
		"password": "password",
	})

	// Set up the taskServiceMock to return nil when DeleteTask is called with any task ID
	taskServiceMock.On("DeleteTask", mock.AnythingOfType("int"), context.Background()).Return(nil)

	// Create a new test request for DELETE /tasks/{taskId}
	req, err := http.NewRequest("DELETE", "/tasks/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a new test response recorder
	rr := httptest.NewRecorder()

	// Call the test request handler with the mock task service and the test request
	router.ServeHTTP(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

}

func TestUpdateTaskHandler(t *testing.T) {
	// Create a new mock task service
	taskServiceMock := &controller.MockTaskService{}

	// Create a new router using Go chi
	r := chi.NewRouter()

	// Register the updateTask handler with the router
	r.Put("/tasks/{taskID}", func(w http.ResponseWriter, r *http.Request) {
		taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
		if err != nil {
			http.Error(w, "invalid task ID", http.StatusBadRequest)
			return
		}
		// Parse the user data from the request body
		var taskData models.Task
		contentLength := r.Header.Get("Content-Length")
		if contentLength == "" {
			http.Error(w, "missing Content-Length header", http.StatusBadRequest)
			return
		}
		// Parse the user data from the request body
		err = json.NewDecoder(r.Body).Decode(&taskData)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		tokenStr := r.Context().Value("token").(string)
		if tokenStr == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if err != nil {
			fmt.Println(err)
			render.Render(w, r, ErrorRenderer(err))
			return
		}
		fmt.Printf("Type of token: %T\n", token)
		task, err := taskServiceMock.UpdateTask(taskID, taskData, context.Background())
		if err != nil {
			if err == repository.ErrNoMatch {
				render.Render(w, r, ErrorRenderer(fmt.Errorf("you are not the manager")))
			} else {
				render.Render(w, r, ServerErrorRenderer(err))
			}
			return
		}
		jsonBytes, err := json.Marshal(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	// Create a new request with a test JWT token
	_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
		"email":    "test@example.com",
		"password": "password",
	})

	// Create a new test task object
	taskData := models.Task{
		Name:           "Test Task",
		Description:    "This is a test task",
		StartDate:      time.Now().Local(),
		EndDate:        time.Now().Add(time.Hour * 24).Local(),
		Status:         null.NewString("Open", true),
		AuthorID:       1,
		TaskCategoryID: 1,
	}
	taskDataJSON, _ := json.Marshal(taskData)
	req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewReader(taskDataJSON))
	if err != nil {
		t.Errorf("error creating test request: %s", err)
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(taskDataJSON)))
	req.ContentLength = int64(len(taskDataJSON))

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))

	// Add the token string to the request context
	ctx := context.WithValue(req.Context(), "token", tokenStr)

	// Create a new request with the updated context
	req = req.WithContext(ctx)

	// Set up the mock task service expectation
	expectedTask := models.Task{
		ID:             1,
		Name:           taskData.Name,
		Description:    taskData.Description,
		StartDate:      taskData.StartDate,
		EndDate:        taskData.EndDate,
		Status:         taskData.Status,
		AuthorID:       taskData.AuthorID,
		TaskCategoryID: taskData.TaskCategoryID,
	}
	taskServiceMock.On("UpdateTask", 1, taskData, context.Background()).Return(expectedTask, nil)

	// Use the httptest package to send the test request to the router
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response body contains the expected task data
	expectedJSON, _ := json.Marshal(expectedTask)
	if body := strings.TrimRight(rr.Body.String(), "\n\t"); body != string(expectedJSON) {
		t.Errorf("handler returned unexpected body: got %v want %v", body, string(expectedJSON))
	}

	// Check that the mock task service expectation was met
	taskServiceMock.AssertExpectations(t)
}

func TestLockTaskHandler(t *testing.T) {
	// Create a new instance of the mock task service
	mockTaskService := &controller.MockTaskService{}

	// Create a new router
	r := chi.NewRouter()

	// Set up the route for the lockTask handler
	r.Put("/tasks/{taskID}/lock", func(w http.ResponseWriter, r *http.Request) {
		taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
		if err != nil {
			http.Error(w, "invalid task ID", http.StatusBadRequest)
			return
		}
		tokenStr := r.Context().Value("token").(string)
		if tokenStr == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if token == nil {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		if err != nil {
			fmt.Println(err)
			render.Render(w, r, ErrorRenderer(err))
			return
		}
		err = mockTaskService.LockTask(taskID, context.Background())
		if err != nil {
			if err == repository.ErrNoMatch {
				render.Render(w, r, ErrNotFound)
			} else {
				render.Render(w, r, ServerErrorRenderer(err))
			}
			return
		}
	})

	// Define the test case
	testCases := []struct {
		name           string
		taskID         int
		expectedStatus int
		expectedError  error
	}{
		{
			name:           "Success - Task locked",
			taskID:         1,
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid task ID",
			taskID:         0,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  errors.New("invalid task ID"),
		},
		{
			name:           "Error - Task not found",
			taskID:         2,
			expectedStatus: http.StatusNotFound,
			expectedError:  repository.ErrNoMatch,
		},
	}
	// Create a new request with a test JWT token
	_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
		"email":    "test@example.com",
		"password": "password",
	})
	// Loop through each test case
	for _, tc := range testCases {
		// Reset the mock
		mockTaskService.On("LockTask", tc.taskID, context.Background()).Return(tc.expectedError)

		// Create a new request
		req := httptest.NewRequest("PUT", fmt.Sprintf("/tasks/%d/lock", tc.taskID), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))

		// Add the token string to the request context
		ctx := context.WithValue(req.Context(), "token", tokenStr)

		// Create a new request with the updated context
		req = req.WithContext(ctx)
		// Create a new response recorder
		recorder := httptest.NewRecorder()

		// Call the handler function
		r.ServeHTTP(recorder, req)

		// Check the response status code
		if recorder.Result().StatusCode != tc.expectedStatus {
			t.Errorf("%s: expected status code %d, but got %d", tc.name, tc.expectedStatus, recorder.Result().StatusCode)
		}

		// Check the mock function was called
		mockTaskService.AssertCalled(t, "LockTask", tc.taskID, context.Background())
	}
}
func TestUnLockTaskHandler(t *testing.T) {
	// Create a new instance of the mock task service
	mockTaskService := &controller.MockTaskService{}

	// Create a new router
	r := chi.NewRouter()

	// Set up the route for the unLockTask handler
	r.Put("/tasks/{taskID}/unlock", func(w http.ResponseWriter, r *http.Request) {
		taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
		if err != nil {
			http.Error(w, "invalid task ID", http.StatusBadRequest)
			return
		}
		tokenStr := r.Context().Value("token").(string)
		if tokenStr == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if err != nil {
			fmt.Println(err)
			render.Render(w, r, ErrorRenderer(err))
			return
		}
		if token == nil {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		err = mockTaskService.UnLockTask(taskID, context.Background())
		if err != nil {
			if err == repository.ErrNoMatch {
				render.Render(w, r, ErrNotFound)
			} else {
				render.Render(w, r, ServerErrorRenderer(err))
			}
			return
		}
	})

	// Define the test cases
	testCases := []struct {
		name           string
		taskID         int
		tokenStr       string
		expectedStatus int
		expectedError  error
	}{
		{
			name:           "Success - Task unlocked",
			taskID:         1,
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid task ID",
			taskID:         0,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  fmt.Errorf("invalid task ID"),
		},

		{
			name:           "Error - Task not found",
			taskID:         2,
			expectedStatus: http.StatusNotFound,
			expectedError:  repository.ErrNoMatch,
		},
	}
	// Create a new request with a test JWT token
	_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
		"email":    "test@example.com",
		"password": "password",
	})

	// Loop through the test cases
	for _, tc := range testCases {
		// Create a new request with the test token
		req := httptest.NewRequest("PUT", fmt.Sprintf("/tasks/%d/unlock", tc.taskID), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.tokenStr))

		// Add the token string to the request context
		ctx := context.WithValue(req.Context(), "token", tokenStr)

		// Create a new request with the updated context
		req = req.WithContext(ctx)

		// Set up the mock behavior
		mockTaskService.On("UnLockTask", tc.taskID, context.Background()).Return(tc.expectedError)

		// Call the handler function
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, req)

		// Check the response status code
		if recorder.Result().StatusCode != tc.expectedStatus {
			t.Errorf("%s: expected status code %d, but got %d", tc.name, tc.expectedStatus, recorder.Result().StatusCode)
		}

		// Check the mock function was called
		mockTaskService.AssertCalled(t, "UnLockTask", tc.taskID, context.Background())
	}
}

func TestGetTaskCategoryOfTask(t *testing.T) {
	// Define the mock task category service
	taskServiceMock := &controller.MockTaskService{}

	// Define the test cases
	testCases := []struct {
		name           string
		taskID         int
		mockTaskCat    *models.TaskCategory
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			taskID:         1,
			mockTaskCat:    &models.TaskCategory{ID: 1, Name: "Category 1"},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"Category 1"}`,
		},
		{
			name:           "Task Category Not Found",
			taskID:         2,
			mockTaskCat:    &models.TaskCategory{},
			mockErr:        repository.ErrNoMatch,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"status_text":"","message":"Resource not found"}`,
		},
		{
			name:           "Invalid Task ID",
			taskID:         0,
			mockTaskCat:    &models.TaskCategory{},
			mockErr:        fmt.Errorf("invalid task ID"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status_text":"Bad request","message":"invalid task ID"}`,
		},
	}

	// Loop through the test cases and run each one
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the taskCategoryServiceMock to return the mock task category and error for this test case
			taskServiceMock.On("GetTaskCategoryOfTask", tt.taskID, context.Background()).Return(tt.mockTaskCat, tt.mockErr)

			// Create a new test request for GET /task-categories/{id}
			req, err := http.NewRequest("GET", fmt.Sprintf("/tasks/%d/get-task-category", tt.taskID), nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a new test response recorder
			rr := httptest.NewRecorder()

			// Create a new chi router and add a route handler for GET /task-categories/{id}
			router := chi.NewRouter()
			router.Get("/tasks/{taskID}/get-task-category", func(w http.ResponseWriter, r *http.Request) {
				// Get the task category ID from the URL parameter
				taskID := chi.URLParam(r, "taskID")

				// Parse the task category ID as an integer
				id, err := strconv.Atoi(taskID)
				if err != nil {
					render.Render(w, r, ErrorRenderer(err))
					return
				}

				// Call the task category service to get the task category with the specified ID
				taskCategory, err := taskServiceMock.GetTaskCategoryOfTask(id, context.Background())
				if err != nil {
					if err == repository.ErrNoMatch {
						render.Render(w, r, ErrNotFound)
					} else {
						render.Render(w, r, ErrorRenderer(err))
					}
					return
				}

				// Render the task category as JSON and send it as the response body
				jsonBytes, err := json.Marshal(taskCategory)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			})

			// Call the test request handler with the mock task category service and the test request
			router.ServeHTTP(rr, req)

			// Check that the response status code matches the expected status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check that the response body matches the expected body
			response := strings.TrimRight(rr.Body.String(), "\n\t\r")
			if response != tt.expectedBody {
				t.Errorf("Handler returned unexpected body: got %q want %q", response, tt.expectedBody)
			}
		})
	}
}

// func TestImportTaskCSV(t *testing.T) {
// 	// Create a new router
// 	r := chi.NewRouter()

// 	// Create a new mock task service
// 	mockTaskService := controller.MockTaskService{}
// 	mockUserTaskDetailService := controller.MockUserTaskDetailService{}

// 	// Set up the route handler with the mock task service
// 	r.Post("/tasks/csv", func(w http.ResponseWriter, r *http.Request) {
// 		// Parse the form data
// 		err := r.ParseForm()
// 		if err != nil {
// 			render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
// 			return
// 		}

// 		// Get the path parameter from the form data
// 		path := r.PostForm.Get("path")

// 		// Import the task data from the CSV file
// 		taskList, err := mockTaskService.ImportTaskDataFromCSV(path)
// 		if err != nil {
// 			render.Render(w, r, ErrorRenderer(err))
// 			return
// 		}

// 		// Get the authentication token
// 		token := GetToken(r, tokenAuth)
// 		if token == nil {
// 			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
// 			return
// 		}
// 		// Add each task to the database and assign it to the author
// 		for _, task := range taskList {
// 			if err := mockTaskService.AddTask(&task, context.Background()); err != nil {
// 				render.Render(w, r, ErrorRenderer(err))
// 				return
// 			}
// 			if err = mockUserTaskDetailService.AddUserToTask(task.AuthorID, task.ID, context.Background()); err != nil {
// 				render.Render(w, r, ErrorRenderer(err))
// 				return
// 			}
// 		}

// 		// Render the task list as the response
// 		jsonBytes, err := json.Marshal(taskList)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		w.Write(jsonBytes)
// 	})

// 	// Create a new test request with a path parameter
// 	reqBody := fmt.Sprintf("path=%s", "./data/task.csv")
// 	req := httptest.NewRequest("POST", "/tasks/csv", strings.NewReader(reqBody))

// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	// Create a new test response recorder
// 	rr := httptest.NewRecorder()

// 	// Set up the expected task list returned by the mock task service
// 	expectedTaskList := []models.Task{
// 		{
// 			ID:             1,
// 			Name:           "Task 1",
// 			Description:    "Description for Task 1",
// 			StartDate:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
// 			EndDate:        time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
// 			Status:         null.NewString("In Progress", true),
// 			AuthorID:       1,
// 			CreatedAt:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
// 			UpdatedAt:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
// 			TaskCategoryID: 1,
// 		},
// 		{
// 			ID:             2,
// 			Name:           "Task 2",
// 			Description:    "Description for Task 2",
// 			StartDate:      time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
// 			EndDate:        time.Date(2022, 2, 28, 0, 0, 0, 0, time.UTC),
// 			Status:         null.NewString("Completed", true),
// 			AuthorID:       2,
// 			CreatedAt:      time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
// 			UpdatedAt:      time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
// 			TaskCategoryID: 2,
// 		},
// 	}

// 	// Set up the expected calls to the mock task service
// 	mockTaskService.On("ImportTaskDataFromCSV", "./data/task.csv").Return(expectedTaskList, nil)

// 	mockTaskService.On("AddTask", &expectedTaskList[0]).Return(nil)
// 	mockTaskService.On("AddTask", &expectedTaskList[1]).Return(nil)

// 	mockTaskService.On("AddUserToTask", 1, 1).Return(nil)
// 	mockTaskService.On("AddUserToTask", 2, 2).Return(nil)
// 	// Perform the test request
// 	r.ServeHTTP(rr, req)

// 	// Check the response status code
// 	if rr.Code != http.StatusOK {
// 		t.Errorf("Expected status code %d; got %d", http.StatusOK, rr.Code)
// 	}

// 	// Check the response body
// 	var responseTaskList []models.Task
// 	err := json.NewDecoder(rr.Body).Decode(&responseTaskList)
// 	if err != nil {
// 		t.Errorf("Error decoding response body: %s", err)
// 	}
// 	if !reflect.DeepEqual(responseTaskList, expectedTaskList) {
// 		t.Errorf("Response body does not match expected task list: expected %+v, got %+v", expectedTaskList, responseTaskList)
// 	}

// 	// Check the calls to the mock task service
// 	mockTaskService.AssertExpectations(t)
// }
