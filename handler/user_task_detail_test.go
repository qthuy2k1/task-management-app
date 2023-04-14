package handler

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"strconv"
// 	"strings"
// 	"testing"

// 	"github.com/go-chi/chi/v5"
// 	"github.com/go-chi/render"
// 	"github.com/qthuy2k1/task-management-app/db"
// )

// func TestAddUserTaskDetailHandler(t *testing.T) {
// 	// Create a new instance of the mock user task detail service
// 	mockUserTaskDetailService := &db.MockUserTaskDetailService{}

// 	// Create a new router
// 	r := chi.NewRouter()

// 	// Set up the route for the createUserTaskDetail handler
// 	r.Post("/tasks/{taskID}/add-user", func(w http.ResponseWriter, r *http.Request) {
// 		err := r.ParseForm()
// 		if err != nil {
// 			render.Render(w, r, ServerErrorRenderer(fmt.Errorf("failed to parse form data")))
// 		}
// 		userID, err := strconv.Atoi(r.PostForm.Get("id"))
// 		if err != nil {
// 			render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid user id")))
// 			return
// 		}
// 		taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
// 		if err != nil {
// 			render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task id")))
// 			return
// 		}

// 		tokenStr := r.Context().Value("token").(string)
// 		if tokenStr == "" {
// 			render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
// 			return
// 		}
// 		token, err := tokenAuth.Decode(tokenStr)
// 		if err != nil {
// 			render.Render(w, r, ErrorRenderer(err))
// 			return
// 		}

// 		err = mockUserTaskDetailService.AddUserToTask(userID, taskID, r, tokenAuth, token)
// 		if err != nil {
// 			render.Render(w, r, ErrorRenderer(err))
// 			return
// 		}
// 	})

// 	// Define the test cases as a slice of structs
// 	testCases := []struct {
// 		name           string
// 		userID         int
// 		taskID         int
// 		isManager      bool
// 		token          string
// 		expectedStatus int
// 		expectedError  error
// 	}{
// 		{
// 			name:           "Success - User added to task",
// 			userID:         1,
// 			taskID:         1,
// 			isManager:      true,
// 			expectedStatus: http.StatusOK,
// 			expectedError:  nil,
// 		},
// 		{
// 			name:           "Invalid user ID",
// 			userID:         0,
// 			taskID:         1,
// 			isManager:      true,
// 			expectedStatus: http.StatusBadRequest,
// 			expectedError:  fmt.Errorf("invalid user id"),
// 		},
// 		// Add more test cases here as needed
// 	}

// 	for _, tc := range testCases {
// 		// Create a new request with the test JWT token
// 		_, tokenStr, _ := tokenAuth.Encode(map[string]interface{}{
// 			"email":    "test@example.com",
// 			"password": "password",
// 		})

// 		// Set up the mock service to return the expected error
// 		mockUserTaskDetailService.On("AddUserToTask", tc.userID, tc.taskID).Return(tc.expectedError)

// 		// Create a new request with the test user ID and task ID
// 		formData := url.Values{}
// 		formData.Set("id", strconv.Itoa(tc.userID))
// 		req := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d/add-user", tc.taskID), strings.NewReader(formData.Encode()))
// 		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))

// 		// Add the token string to the context
// 		ctx := context.WithValue(req.Context(), "token", tokenStr)
// 		req = req.WithContext(ctx)

// 		// Create a new recorder for the response
// 		recorder := httptest.NewRecorder()

// 		// Call the handler with the test request and recorder
// 		r.ServeHTTP(recorder, req)

// 		// Check the response status code
// 		if recorder.Result().StatusCode != tc.expectedStatus {
// 			t.Errorf("%s: expected status code %d but got %d", tc.name, tc.expectedStatus, recorder.Result().StatusCode)
// 		}

// 		// Check the mock service was called with the expected arguments
// 		mockUserTaskDetailService.AssertCalled(t, "AddUserToTask", tc.userID, tc.taskID)

// 		// Reset the mock service
// 		mockUserTaskDetailService.AssertExpectations(t)
// 	}
// }
