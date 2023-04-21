package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/internal/handlers"
	mockControllers "github.com/qthuy2k1/task-management-app/internal/mocks/controllers"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func TestGetUserByIDHandler(t *testing.T) {
	// Create a mock user service
	mockUserService := &mockControllers.MockUserService{}

	// Set up the test cases
	testCases := []struct {
		userID      int
		expected    *models.User
		expectedErr error
	}{
		{
			userID: 1,
			expected: &models.User{
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
			expected:    &models.User{},
			expectedErr: repositories.ErrNoMatch,
		},
		{
			userID: 24,
			expected: &models.User{
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
			expected: &models.User{
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
			expected:    &models.User{},
			expectedErr: repositories.ErrNoMatch,
		},
		{
			userID:      -1,
			expected:    &models.User{},
			expectedErr: handlers.ErrNotFound.Err,
		},
	}

	// Set up the mock user service to return the expected values for each test case
	for _, tc := range testCases {
		mockUserService.On("GetUserByID", tc.userID, context.Background()).Return(tc.expected, tc.expectedErr)
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
		user, err := mockUserService.GetUserByID(userID, context.Background())
		if err != nil {
			if err == repositories.ErrNoMatch {
				http.Error(w, handlers.ErrNotFound.Message, http.StatusNotFound)
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
		} else if tc.expectedErr == repositories.ErrNoMatch {
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
	mockUserList := models.UserSlice{
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
	}

	mockUserService := &mockControllers.MockUserService{}

	// Define the expected return values for the mock GetAllUsers function
	mockUserService.On("GetAllUsers", context.Background()).Return(mockUserList, nil)

	// Create a new instance of the Chi router and register the GetAllUsers route
	r := chi.NewRouter()
	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		userlist, err := mockUserService.GetAllUsers(context.Background())
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
	mockUserService := &mockControllers.MockUserService{}

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
		updatedUser, err := mockUserService.UpdateUser(userID, userData, context.Background())
		if err != nil {
			if err == repositories.ErrNoMatch {
				http.Error(w, err.Error(), handlers.ErrNotFound.StatusCode)
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
		expectedUser  *models.User
		expectedError error
	}{
		{
			userID: 1,
			userData: models.User{
				Name:  "Alice",
				Email: "alice@example.com",
			},
			expectedUser: &models.User{
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
			expectedUser:  &models.User{},
			expectedError: repositories.ErrNoMatch,
		},
	}

	// Set up the mock user service to return the expected values for each test case
	for _, tc := range testCases {
		mockUserService.On("UpdateUser", tc.userID, tc.userData, context.Background()).Return(tc.expectedUser, tc.expectedError)
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
			var updatedUser *models.User
			err := json.NewDecoder(w.Body).Decode(&updatedUser)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedUser, updatedUser)
		} else if tc.expectedError == repositories.ErrNoMatch {
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
	// Set up the mock user service
	mockUserService := &mockControllers.MockUserService{}

	// Set up the expectations for each test case
	mockUserService.On("UpdateRole", 1, "manager", context.Background()).Return(&models.User{
		ID:    1,
		Name:  "Alice",
		Email: "alice@example.com",
		Role:  "manager",
	}, nil)
	mockUserService.On("UpdateRole", 2, "user", context.Background()).Return(&models.User{
		ID:    2,
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  "user",
	}, nil)
	mockUserService.On("UpdateRole", 3, "adad", context.Background()).Return(&models.User{}, repositories.ErrNoMatch)
	mockUserService.On("UpdateRole", 4, "", context.Background()).Return(&models.User{}, invalidRoleErr)

	// Set up the test cases
	testCases := []struct {
		userID        int
		role          string
		expectedUser  *models.User
		expectedError error
	}{
		{
			userID: 1,
			role:   "manager",
			expectedUser: &models.User{
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
			expectedUser: &models.User{
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
			expectedUser:  &models.User{},
			expectedError: repositories.ErrNoMatch,
		},
		{
			userID:        4,
			role:          "",
			expectedUser:  &models.User{},
			expectedError: invalidRoleErr,
		},
	}

	r := chi.NewRouter()
	// Add a route for updating a user
	r.Patch("/users/{userID}/update-role", func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		// Parse the user ID from the URL path
		userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
		if err != nil {
			http.Error(w, "cannot get the user id", http.StatusBadRequest)
			return
		}

		// Call the UpdateUser function on the mock user service
		updatedUser, err := mockUserService.UpdateRole(userID, role, context.Background())
		if role == "" {
			err = invalidRoleErr
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			if err == repositories.ErrNoMatch {
				http.Error(w, err.Error(), handlers.ErrNotFound.StatusCode)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Return the updated user as JSON in the response body
		json.NewEncoder(w).Encode(updatedUser)
	})

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
			var updatedUser *models.User
			err := json.NewDecoder(w.Body).Decode(&updatedUser)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedUser, updatedUser)
			assert.Equal(t, tc.role, tc.expectedUser.Role)
		} else if errors.Is(tc.expectedError, repositories.ErrNoMatch) {
			assert.Equal(t, http.StatusNotFound, w.Code)
		} else if errors.Is(tc.expectedError, invalidRoleErr) {
			assert.Equal(t, http.StatusBadRequest, w.Code)
		} else {
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}
	}

	// Assert that all expectations were met
	mockUserService.AssertExpectations(t)

}

func TestSignUpHandler(t *testing.T) {
	testCases := []struct {
		name         string
		formData     url.Values
		expectedCode int
		expectedUser *models.User
	}{
		{
			name: "valid signup",
			formData: url.Values{
				"name":     {"John Doe"},
				"email":    {"johndoe@example.com"},
				"password": {"password123"},
			},
			expectedCode: http.StatusOK,
			expectedUser: &models.User{
				Name:     "John Doe",
				Email:    "johndoe@example.com",
				Password: "password123",
				Role:     "user",
			},
		},
		{
			name: "missing name",
			formData: url.Values{
				"email":    {"johndoe@example.com"},
				"password": {"password123"},
			},
			expectedCode: http.StatusBadRequest,
			expectedUser: nil,
		},
		// add more test cases here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock user service
			mockUserService := &mockControllers.MockUserService{}

			// Set up the mock user service to return nil error for AddUser
			mockUserService.On("AddUser", tc.expectedUser, context.Background()).Return(nil)

			// Create a new router with the SignUp handler
			router := chi.NewRouter()
			router.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
				// Call the SignUp handler with the mock user service
				err := r.ParseForm()
				if err != nil {
					render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("failed to parse form data")))
					return
				}

				user := models.User{}
				user.Name = r.PostForm.Get("name")
				user.Email = r.PostForm.Get("email")
				user.Password = r.PostForm.Get("password")
				user.Role = "user"

				if user.Email == "" || user.Name == "" || user.Password == "" {
					render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("missing name, email or password")))
					return
				}

				// // Validate email
				// if !db.IsValidEmail(user.Email) {
				// 	render.Render(w, r, ErrorRenderer(fmt.Errorf("your email is not valid, please provide a valid email")))
				// 	return
				// }

				// // Validate password
				// // The password must contain at least 6 characters
				// if !controller.IsValidPassword(user.Password) {
				// 	render.Render(w, r, ErrorRenderer(fmt.Errorf("your password is not valid, please provide a password that contains at least 6 characters")))
				// 	return
				// }

				// Add user to database
				err = mockUserService.AddUser(&user, context.Background())
				if err != nil {
					render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("failed to add user to database")))
					return
				}

				// Generate JWT token for user
				token := handlers.MakeToken(user.Email, user.Password)

				// Set JWT token as cookie
				http.SetCookie(w, &http.Cookie{
					HttpOnly: true,
					Expires:  time.Now().Add(7 * 24 * time.Hour),
					SameSite: http.SameSiteLaxMode,
					// Uncomment below for HTTPS:
					// Secure: true,
					Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
					Value: token,
				})

			})

			// Create a new signup request with mock form data
			signupRequest, err := http.NewRequest("POST", "/signup", strings.NewReader(tc.formData.Encode()))
			if err != nil {
				t.Fatalf("Failed to create signup request: %v", err)
			}
			signupRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			// Call the signup request using the router and check for successful signup response
			signupResponse := httptest.NewRecorder()
			router.ServeHTTP(signupResponse, signupRequest)

			if signupResponse.Code != tc.expectedCode {
				t.Errorf("Expected status code %d, but got %d", tc.expectedCode, signupResponse.Code)
			}

			// Verify that the AddUser method was called on the mock user service with the correct user object
			if tc.expectedUser != nil {
				mockUserService.AssertCalled(t, "AddUser", tc.expectedUser, context.Background())
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	// Create a new mock user service instance for each test case
	mockUserService := &mockControllers.MockUserService{}
	// Create a new router with the login handler
	router := chi.NewRouter()
	router.Post("/login",
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Failed to parse form data.", http.StatusInternalServerError)
				return
			}

			email := r.PostForm.Get("email")
			password := r.PostForm.Get("password")

			// Define a regular expression for validating email addresses
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

			// Use the MatchString method to check if the email address matches the regular expression
			validEmail := emailRegex.MatchString(email)
			if email == "" || !validEmail {
				render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("your email is not valid, please provide a valid email")))
				return
			}

			// Validate password
			// The password must contain at least 6 characters
			if password == "" || len(password) < 6 {
				render.Render(w, r, handlers.ErrorRenderer(fmt.Errorf("your password is not valid, please provide a valid password that contains at least 6 characters")))
				return
			}

			token := handlers.MakeToken(email, password)

			// Set the JWT token as a cookie in the response
			http.SetCookie(w, &http.Cookie{
				HttpOnly: true,
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				SameSite: http.SameSiteLaxMode,
				// Uncomment below for HTTPS:
				// Secure: true,
				Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
				Value: token,
			})

			// Write the response with a success message
			w.Write([]byte(fmt.Sprintf(`Login successful, your email is: %s`, email)))

			// }
		})
	// Define a slice of test cases
	testCases := []struct {
		name               string
		email              string
		password           string
		expectedStatusCode int
		expectedBody       string
		isValidEmail       bool
		isValidPassword    bool
	}{
		{
			name:               "Login with valid email and password",
			email:              "test@example.com",
			password:           "password",
			expectedStatusCode: http.StatusOK,
			expectedBody:       `Login successful, your email is: test@example.com`,
		},
		{
			name:               "Login with missing email",
			email:              "",
			password:           "password",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `your email is not valid, please provide a valid email`,
		},
		{
			name:               "Login with missing password",
			email:              "test@example.com",
			password:           "",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `your password is not valid, please provide a valid password that contains at least 6 characters`,
		},
		{
			name:               "Login with invalid email",
			email:              "invalid-email",
			password:           "password",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `your email is not valid, please provide a valid email`,
		},
		{
			name:               "Login with weak password",
			email:              "test@example.com",
			password:           "weak",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `your password is not valid, please provide a valid password that contains at least 6 characters`,
		},
	}
	// Loop through the test cases
	for _, tc := range testCases {
		// Create a new test request with the email and password form data
		reqBody := fmt.Sprintf("email=%s&password=%s", tc.email, tc.password)
		req, err := http.NewRequest("POST", "/login", strings.NewReader(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Create a new test response recorder
		recorder := httptest.NewRecorder()

		// Call the router with the test request and response recorder
		router.ServeHTTP(recorder, req)

		// Check if the response status code is as expected
		if recorder.Code != tc.expectedStatusCode {
			t.Errorf("%s: Expected response status code %d, but got %d", tc.name, tc.expectedStatusCode, recorder.Code)
			continue
		}

		// Check if the response body contains the expected message
		if !strings.Contains(recorder.Body.String(), tc.expectedBody) {
			t.Errorf("%s: Expected response body to contain %q, but got %q", tc.name, tc.expectedBody, recorder.Body.String())
			continue
		}

		// Assert that the mock user service was called as expected
		mockUserService.AssertExpectations(t)
	}
}

func TestLogoutHandler(t *testing.T) {
	// Create a new router with the logout handler
	router := chi.NewRouter()
	router.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			MaxAge:   -1, // Delete the cookie.
			SameSite: http.SameSiteLaxMode,
			// Uncomment below for HTTPS:
			// Secure: true,
			Name:  "jwt",
			Value: "",
		})
		w.Write([]byte(`Logout successful`))
	})

	// Create a new test request to the logout route
	req, err := http.NewRequest("GET", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new test response recorder
	recorder := httptest.NewRecorder()

	// Call the router with the test request and response recorder
	router.ServeHTTP(recorder, req)

	// Check if the response status code is as expected
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected response status code %d, but got %d", http.StatusOK, recorder.Code)
	}

	// Check if the cookie was deleted
	cookie := recorder.Header().Get("Set-Cookie")
	if cookie != "jwt=; Max-Age=0; HttpOnly; SameSite=Lax" {
		t.Errorf("Expected cookie to be deleted, but got %s", cookie)
	}
}

func TestGetUsersManager(t *testing.T) {
	// Create a mock user list
	mockUserList := models.UserSlice{
		{
			ID:       1,
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Password: "password1",
			Role:     "manager",
		},
		{
			ID:       2,
			Name:     "Jane Smith",
			Email:    "jane.smith@example.com",
			Password: "password2",
			Role:     "manager",
		},
	}

	mockUserService := &mockControllers.MockUserService{}

	// Define the expected return values for the mock GetAllUsers function
	mockUserService.On("GetUsersManager", context.Background()).Return(mockUserList, nil)

	// Create a new instance of the Chi router and register the GetAllUsers route
	r := chi.NewRouter()
	r.Get("/users/managers", func(w http.ResponseWriter, r *http.Request) {
		userlist, err := mockUserService.GetUsersManager(context.Background())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userlist)
	})

	// Create a new HTTP request that targets the GetAllUsers route
	req, err := http.NewRequest("GET", "/users/managers", nil)
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
