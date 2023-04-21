package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/internal/handlers"
	mockControllers "github.com/qthuy2k1/task-management-app/internal/mocks/controllers"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repositories"
	"github.com/stretchr/testify/mock"
)

func TestGetAllTaskCategories(t *testing.T) {
	// Create a new mock task category service
	taskCategoryServiceMock := &mockControllers.MockTaskCategoryService{}

	r := chi.NewRouter()

	// Set up the mock task category service to return a list of task categories
	taskCategoryList := models.TaskCategorySlice{
		{
			ID:   1,
			Name: "Category 1",
		},
		{
			ID:   2,
			Name: "Category 2",
		},
	}
	taskCategoryServiceMock.On("GetAllTaskCategories", context.Background()).Return(taskCategoryList, nil)

	// Create a new test request
	req, err := http.NewRequest("GET", "/taskCategories", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set up the request context with a valid JWT token

	// Create a new test response recorder
	rr := httptest.NewRecorder()

	// Call the getAllTaskCategories function with the mock task category service and the test request
	r.Get("/taskCategories", func(w http.ResponseWriter, r *http.Request) {
		// Get the JWT token from the request header
		if err != nil {
			return
			// handle error
		}
		if err != nil {
			http.Error(w, "Invalid JWT token", http.StatusBadRequest)
			return
		}

		// Call the task category service to get the list of task categories
		taskCategoryList, err := taskCategoryServiceMock.GetAllTaskCategories(context.Background())
		if err != nil {
			http.Error(w, "Error retrieving task categories", http.StatusInternalServerError)
			return
		}

		// Render the task category list as JSON and send it as the response body
		render.JSON(w, r, taskCategoryList)
	})

	r.ServeHTTP(rr, req)

	// Assert that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Assert that the response body matches the expected task category list
	expectedBody, _ := json.Marshal(taskCategoryList)
	response := strings.TrimRight(rr.Body.String(), "\n\t\r")
	if response != string(expectedBody) {
		t.Errorf("Handler returned unexpected body: got %q want %q", response, expectedBody)
	}

	// Assert that the GetAllTaskCategories method was called on the mock task category service with the correct arguments
	taskCategoryServiceMock.AssertCalled(t, "GetAllTaskCategories", context.Background())
}
func TestGetTaskCategoryByID(t *testing.T) {
	// Define the mock task category service
	taskCategoryServiceMock := &mockControllers.MockTaskCategoryService{}

	// Define the test cases
	testCases := []struct {
		name           string
		taskCategoryID int
		mockTaskCat    *models.TaskCategory
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			taskCategoryID: 1,
			mockTaskCat:    &models.TaskCategory{ID: 1, Name: "Category 1"},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"Category 1"}`,
		},
		{
			name:           "Task Category Not Found",
			taskCategoryID: 2,
			mockTaskCat:    &models.TaskCategory{},
			mockErr:        repositories.ErrNoMatch,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"status_text":"","message":"Resource not found"}`,
		},
		{
			name:           "Invalid Task Category ID",
			taskCategoryID: 0,
			mockTaskCat:    &models.TaskCategory{},
			mockErr:        fmt.Errorf("invalid task category ID"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status_text":"Bad request","message":"invalid task category ID"}`,
		},
	}

	// Loop through the test cases and run each one
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the taskCategoryServiceMock to return the mock task category and error for this test case
			taskCategoryServiceMock.On("GetTaskCategoryByID", tt.taskCategoryID, context.Background()).Return(tt.mockTaskCat, tt.mockErr)

			// Create a new test request for GET /task-categories/{id}
			req, err := http.NewRequest("GET", fmt.Sprintf("/task-categories/%d", tt.taskCategoryID), nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a new test response recorder
			rr := httptest.NewRecorder()

			// Create a new chi router and add a route handler for GET /task-categories/{id}
			router := chi.NewRouter()
			router.Get("/task-categories/{taskCategoryID}", func(w http.ResponseWriter, r *http.Request) {
				// Get the task category ID from the URL parameter
				taskCategoryID := chi.URLParam(r, "taskCategoryID")

				// Parse the task category ID as an integer
				id, err := strconv.Atoi(taskCategoryID)
				if err != nil {
					render.Render(w, r, handlers.ErrorRenderer(err))
					return
				}

				// Call the task category service to get the task category with the specified ID
				taskCategory, err := taskCategoryServiceMock.GetTaskCategoryByID(id, context.Background())
				if err != nil {
					if err == repositories.ErrNoMatch {
						render.Render(w, r, handlers.ErrNotFound)
					} else {
						render.Render(w, r, handlers.ErrorRenderer(err))
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

func TestAddTaskCategory(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		taskCategory   *models.TaskCategory
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid task category",
			taskCategory: &models.TaskCategory{
				ID:   1,
				Name: "Test Task Category",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"Test Task Category"}`,
		},
		{
			name: "Invalid request body",
			taskCategory: &models.TaskCategory{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock task service
			taskCategoryServiceMock := &mockControllers.MockTaskCategoryService{}

			// Create a new test router
			router := chi.NewRouter()

			// Register the createTaskCategory function as a handler for POST requests to /task_categories
			router.Post("/task_categories", func(w http.ResponseWriter, r *http.Request) {
				// Parse the task category data from the request body
				var taskCategoryData models.TaskCategory
				err := json.NewDecoder(r.Body).Decode(&taskCategoryData)
				if err != nil {
					http.Error(w, "invalid request body", http.StatusBadRequest)
					return
				}

				tokenStr := strings.Split(r.Header.Get("Authorization"), " ")[1]
				if tokenStr == "" {
					render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
					return
				}
				token, err := tokenAuth.Decode(tokenStr)
				if err != nil {
					fmt.Println(err)
					render.Render(w, r, handlers.ErrorRenderer(err))
					return
				}
				if token == nil {
					fmt.Println(err)
					render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
					return
				}

				// Call the task service to add the new task category
				err = taskCategoryServiceMock.AddTaskCategory(&taskCategoryData, context.Background())
				if taskCategoryData.Name == "" {
					http.Error(w, "invalid request body", http.StatusBadRequest)
					return
				}
				if err != nil {
					render.Render(w, r, handlers.ErrorRenderer(err))
					return
				}

				// Render the new task category as JSON and send it as the response body
				jsonBytes, err := json.Marshal(taskCategoryData)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			})

			// Set up the taskServiceMock to return nil when AddTaskCategory is called with the mock task category
			taskCategoryServiceMock.On("AddTaskCategory", tc.taskCategory, context.Background()).Return(nil)
			fmt.Println("=======================================================")
			_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
				"email":    "test@example.com",
				"password": "password",
			})
			// Encode the mock task category as JSON and create a new test request for POST /task_categories
			taskCategoryJSON, _ := json.Marshal(tc.taskCategory)
			req, err := http.NewRequest("POST", "/task_categories", bytes.NewReader(taskCategoryJSON))
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

			// Check that the response status code matches the expected status code
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}

			// Check that the response body matches the expected body
			response := strings.TrimRight(rr.Body.String(), "\n\t\r")
			if response != tc.expectedBody {
				t.Errorf("Handler returned unexpected body: got %q want %q", response, tc.expectedBody)
			}
		})
	}
}
func TestDeleteTaskCategory(t *testing.T) {
	// Create a new mock task category service
	taskCategoryServiceMock := &mockControllers.MockTaskCategoryService{}

	// Create a new test router
	router := chi.NewRouter()

	// Register the deleteTaskCategory function as a handler for DELETE requests to /task-categories/{taskCategoryId}
	router.Delete("/task-categories/{taskCategoryId}", func(w http.ResponseWriter, r *http.Request) {
		// Get the task category ID from the URL parameters
		taskCategoryId, err := strconv.Atoi(chi.URLParam(r, "taskCategoryId"))
		if err != nil {
			render.Render(w, r, handlers.ErrBadRequest)
			return
		}

		tokenStr := strings.Split(r.Header.Get("Authorization"), " ")[1]
		if tokenStr == "" {
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if err != nil {
			fmt.Println(err)
			render.Render(w, r, handlers.ErrorRenderer(err))
			return
		}
		if token == nil {
			fmt.Println(err)
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
			return
		}

		// Call the task category service to delete the task category
		err = taskCategoryServiceMock.DeleteTaskCategory(taskCategoryId, context.Background())
		if err != nil {
			if err == repositories.ErrNoMatch {
				render.Render(w, r, handlers.ErrNotFound)
			} else {
				render.Render(w, r, handlers.ServerErrorRenderer(err))
			}
			return
		}

	})
	_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
		"email":    "test@example.com",
		"password": "password",
	})

	// Set up the taskCategoryServiceMock to return nil when DeleteTaskCategory is called with any task category ID
	taskCategoryServiceMock.On("DeleteTaskCategory", mock.AnythingOfType("int"), context.Background()).Return(nil)

	// Create a new test request for DELETE /task-categories/{taskCategoryId}
	req, err := http.NewRequest("DELETE", "/task-categories/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a new test response recorder
	rr := httptest.NewRecorder()

	// Call the test request handler with the mock task category service and the test request
	router.ServeHTTP(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the DeleteTaskCategory method was called on the mock task category service with the correct args
	taskCategoryServiceMock.AssertCalled(t, "DeleteTaskCategory", 1, context.Background())
}

func TestUpdateTaskCategoryHandler(t *testing.T) {
	// Create a new mock task category service
	taskCategoryServiceMock := &mockControllers.MockTaskCategoryService{}

	// Create a new router using Go chi
	r := chi.NewRouter()

	// Register the updateTaskCategory handler with the router
	r.Put("/task-categories/{taskCategoryID}", func(w http.ResponseWriter, r *http.Request) {
		taskCategoryID, err := strconv.Atoi(chi.URLParam(r, "taskCategoryID"))
		if err != nil {
			http.Error(w, "invalid task category ID", http.StatusBadRequest)
			return
		}
		// Parse the task category data from the request body
		var taskCategoryData models.TaskCategory
		contentLength := r.Header.Get("Content-Length")
		if contentLength == "" {
			http.Error(w, "missing Content-Length header", http.StatusBadRequest)
			return
		}
		// Parse the task category data from the request body
		err = json.NewDecoder(r.Body).Decode(&taskCategoryData)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		tokenStr := r.Context().Value("token").(string)
		if tokenStr == "" {
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		token, err := tokenAuth.Decode(tokenStr)
		if err != nil {
			fmt.Println(err)
			render.Render(w, r, handlers.ErrorRenderer(err))
			return
		}
		if token == nil {
			render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("no token found")))
			return
		}
		taskCategory, err := taskCategoryServiceMock.UpdateTaskCategory(taskCategoryID, taskCategoryData, context.Background())
		if err != nil {
			if err == repositories.ErrNoMatch {
				render.Render(w, r, handlers.ErrNotFound)
			} else {
				render.Render(w, r, handlers.ServerErrorRenderer(err))
			}
			return
		}
		jsonBytes, err := json.Marshal(taskCategory)
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

	// Create a new test task category object
	taskCategoryData := models.TaskCategory{Name: "Test Task Category"}
	taskCategoryDataJSON, _ := json.Marshal(taskCategoryData)
	req, err := http.NewRequest("PUT", "/task-categories/1", bytes.NewReader(taskCategoryDataJSON))
	if err != nil {
		t.Errorf("error creating test request: %s", err)
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(taskCategoryDataJSON)))
	req.ContentLength = int64(len(taskCategoryDataJSON))

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))

	// Add the token string to the request context
	ctx := context.WithValue(req.Context(), "token", tokenStr)

	// Create a new request with the updated context
	req = req.WithContext(ctx)

	// Set up the mock task category service expectation
	expectedTaskCategory := &models.TaskCategory{ID: 1, Name: taskCategoryData.Name}
	taskCategoryServiceMock.On("UpdateTaskCategory", 1, taskCategoryData, context.Background()).Return(expectedTaskCategory, nil)

	// Use the httptest package to send the test request to the router
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response body contains the expected task category data
	expectedJSON, _ := json.Marshal(expectedTaskCategory)
	if body := strings.TrimRight(rr.Body.String(), "\n\t"); body != string(expectedJSON) {
		t.Errorf("handler returned unexpected body: got %v want %v", body, string(expectedJSON))
	}

	// Check that the mock task category service expectation was met
	taskCategoryServiceMock.AssertExpectations(t)
}
