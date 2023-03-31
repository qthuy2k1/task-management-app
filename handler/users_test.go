package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/qthuy2k1/task-management-app/db"
	"github.com/qthuy2k1/task-management-app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserByIDHandler(t *testing.T) {
	// Create a mock user service
	mockUserService := &db.MockUserService{}

	// Set up the test cases
	testCases := []struct {
		userID      int
		expected    models.User
		expectedErr error
	}{
		{
			userID: 1,
			expected: models.User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john.doe@example.com",
				Password: "123123",
				Role:     "user",
			},
			expectedErr: nil,
		},
		{
			userID:      2,
			expected:    models.User{},
			expectedErr: db.ErrNoMatch,
		},
		{
			userID: 24,
			expected: models.User{
				ID:       24,
				Name:     "Thuy",
				Email:    "thuy@abc.om",
				Password: "123511",
				Role:     "manager",
			},
			expectedErr: nil,
		},
		{
			userID: 0024,
			expected: models.User{
				ID:       24,
				Name:     "Thuy",
				Email:    "thuy@abc.om",
				Password: "312912",
				Role:     "manager",
			},
			expectedErr: nil,
		},
		{
			userID:      200,
			expected:    models.User{},
			expectedErr: db.ErrNoMatch,
		},
		{
			userID:      -1,
			expected:    models.User{},
			expectedErr: ErrNotFound.Err,
		},
	}

	// Set up the mock user service to return the expected values for each test case
	for _, tc := range testCases {
		mockUserService.On("GetUserByID", tc.userID).Return(tc.expected, tc.expectedErr)
	}

	// Create a new Go Chi router
	r := chi.NewRouter()

	// Add a GET route to the router that calls the GetUserByID method on the user service
	r.Get("/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		userIDStr := chi.URLParam(r, "userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user, err := mockUserService.GetUserByID(userID)
		if err != nil {
			if err == db.ErrNoMatch {
				http.Error(w, ErrNotFound.Message, http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	// Test each test case
	for _, tc := range testCases {
		// Create a test request to the GET route
		req := httptest.NewRequest("GET", fmt.Sprintf("/users/%d", tc.userID), nil)

		// Create a test response recorder
		w := httptest.NewRecorder()

		// Serve the test request using the router
		r.ServeHTTP(w, req)

		// Check the HTTP response code
		if tc.expectedErr == nil {
			assert.Equal(t, http.StatusOK, w.Code)
			// Check the returned user object
			expectedBody, err := json.Marshal(tc.expected)
			assert.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(string(expectedBody)), strings.TrimSpace(w.Body.String()))
		} else if tc.expectedErr == db.ErrNoMatch {
			assert.Equal(t, http.StatusNotFound, w.Code)
		} else {
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}
	}

	// Assert that all expectations were met
	mockUserService.AssertExpectations(t)
}

func TestGetAllUsers(t *testing.T) {
	// Create a mock user list
	mockUserList := &models.UserList{
		Users: []models.User{
			{
				ID:       1,
				Name:     "John Doe",
				Email:    "john.doe@example.com",
				Password: "password1",
				Role:     "admin",
			},
			{
				ID:       2,
				Name:     "Jane Smith",
				Email:    "jane.smith@example.com",
				Password: "password2",
				Role:     "user",
			},
		},
	}

	mockUserService := &db.MockUserService{}

	// Define the expected return values for the mock GetAllUsers function
	mockUserService.On("GetAllUsers", mock.Anything, mock.Anything).Return(mockUserList, nil)

	// Create a new instance of the Chi router and register the GetAllUsers route
	r := chi.NewRouter()
	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		userlist, err := mockUserService.GetAllUsers(r, nil)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userlist)
	})

	// Create a new HTTP request that targets the GetAllUsers route
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the HTTP handler for the GetAllUsers route with the mock request and response recorder
	r.ServeHTTP(rr, req)

	// Check that the HTTP response status code is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check that the HTTP response body matches the expected user list JSON
	expectedJSON, _ := json.Marshal(mockUserList)
	assert.JSONEq(t, string(expectedJSON), rr.Body.String())
}

func TestUpdateUser(t *testing.T) {
	// Create a mock user service
	mockUserService := &db.MockUserService{}

	// Create a new router
	r := chi.NewRouter()

	// Add a route for updating a user
	r.Put("/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		// Parse the user ID from the URL path
		userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}

		// Parse the user data from the request body
		var userData models.User
		err = json.NewDecoder(r.Body).Decode(&userData)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Call the UpdateUser function on the mock user service
		updatedUser, err := mockUserService.UpdateUser(userID, userData)
		if err != nil {
			if err == db.ErrNoMatch {
				http.Error(w, err.Error(), ErrNotFound.StatusCode)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Return the updated user as JSON in the response body
		json.NewEncoder(w).Encode(updatedUser)
	})

	// Set up the test cases
	testCases := []struct {
		userID        int
		userData      models.User
		expectedUser  models.User
		expectedError error
	}{
		{
			userID: 1,
			userData: models.User{
				Name:  "Alice",
				Email: "alice@example.com",
			},
			expectedUser: models.User{
				ID:    1,
				Name:  "Alice",
				Email: "alice@example.com",
			},
			expectedError: nil,
		},
		{
			userID: 2,
			userData: models.User{
				Name:  "Bob",
				Email: "bob@example.com",
			},
			expectedUser:  models.User{},
			expectedError: db.ErrNoMatch,
		},
	}

	// Set up the mock user service to return the expected values for each test case
	for _, tc := range testCases {
		mockUserService.On("UpdateUser", tc.userID, tc.userData).Return(tc.expectedUser, tc.expectedError)
	}

	// Test each test case
	for _, tc := range testCases {
		// Create a new HTTP request with the user data in the request body
		reqBody, err := json.Marshal(tc.userData)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("PUT", fmt.Sprintf("/users/%d", tc.userID), bytes.NewBuffer(reqBody))

		// Create a new HTTP response recorder
		w := httptest.NewRecorder()

		// Call the router's ServeHTTP method to handle the request
		r.ServeHTTP(w, req)

		// Check the HTTP response code
		if tc.expectedError == nil {
			assert.Equal(t, http.StatusOK, w.Code)
			var updatedUser models.User
			err := json.NewDecoder(w.Body).Decode(&updatedUser)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedUser, updatedUser)
		} else if tc.expectedError == db.ErrNoMatch {
			assert.Equal(t, http.StatusNotFound, w.Code)
		} else {
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}
	}

	// Assert that all expectations were met
	mockUserService.AssertExpectations(t)
}

func TestUpdateRoleUser(t *testing.T) {
	invalidRoleErr := fmt.Errorf("invalid role")
	// Create a mock user service
	mockUserService := &db.MockUserService{}

	// Create a new router
	r := chi.NewRouter()

	// Add a route for updating a user
	r.Patch("/users/{userID}/update-role", func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		if role == "" {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}
		// Parse the user ID from the URL path
		userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
		if err != nil {
			http.Error(w, "cannot get the user id", http.StatusBadRequest)
			return
		}

		// Call the UpdateUser function on the mock user service
		updatedUser, err := mockUserService.UpdateRole(userID, role)
		if err != nil {
			if err == db.ErrNoMatch {
				http.Error(w, err.Error(), ErrNotFound.StatusCode)
			} else if errors.Is(err, invalidRoleErr) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Return the updated user as JSON in the response body
		json.NewEncoder(w).Encode(updatedUser)
	})

	// Set up the test cases
	testCases := []struct {
		userID        int
		role          string
		expectedUser  models.User
		expectedError error
	}{
		{
			userID: 1,
			role:   "manager",
			expectedUser: models.User{
				ID:    1,
				Name:  "Alice",
				Email: "alice@example.com",
				Role:  "manager",
			},
			expectedError: nil,
		},
		{
			userID: 2,
			role:   "user",
			expectedUser: models.User{
				ID:    2,
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  "user",
			},
			expectedError: nil,
		},
		{
			userID:        3,
			role:          "adad",
			expectedUser:  models.User{},
			expectedError: db.ErrNoMatch,
		},
		{
			userID:        4,
			role:          "",
			expectedUser:  models.User{},
			expectedError: invalidRoleErr,
		},
	}

	// Set up the mock user service to return the expected values for each test case
	for _, tc := range testCases {
		mockUserService.On("UpdateRole", tc.userID, tc.role).Return(tc.expectedUser, tc.expectedError)
	}

	// Test each test case
	for _, tc := range testCases {
		// Create a new HTTP request with the user data in the request body
		req := httptest.NewRequest("PATCH", fmt.Sprintf("/users/%d/update-role?role=%s", tc.userID, tc.role), nil)

		// Create a new HTTP response recorder
		w := httptest.NewRecorder()

		// Call the router's ServeHTTP method to handle the request
		r.ServeHTTP(w, req)

		// Check the HTTP response code
		if tc.expectedError == nil {
			assert.Equal(t, http.StatusOK, w.Code)
			var updatedUser models.User
			err := json.NewDecoder(w.Body).Decode(&updatedUser)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedUser, updatedUser)
			assert.Equal(t, tc.role, tc.expectedUser.Role)
		} else if errors.Is(tc.expectedError, db.ErrNoMatch) {
			assert.Equal(t, http.StatusNotFound, w.Code)
		} else if errors.Is(tc.expectedError, invalidRoleErr) {
			assert.Equal(t, http.StatusBadRequest, w.Code)
		} else {
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}
		fmt.Println(tc.userID)
	}

	// Assert that all expectations were met
	mockUserService.AssertExpectations(t)
}
