package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/internal/controller"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
)

type UserHandler struct {
	UserController           *controller.UserController
	UserTaskDetailController *controller.UserTaskDetailController
}

func NewUserHandler(database *repository.Database) *UserHandler {
	userRepository := repository.NewUserRepository(database)
	userController := controller.NewUserController(userRepository)
	userTaskDetailRepository := repository.NewUserTaskDetailRepository(database)
	userTaskDetailController := controller.NewUserTaskDetailController(userTaskDetailRepository)
	return &UserHandler{UserController: userController, UserTaskDetailController: userTaskDetailController}
}

type success struct {
	Status string `json:"status"`
}

func (h *UserHandler) users(router chi.Router) {
	router.Get("/", h.getAllUsers)
	router.Post("/change-password", h.changeUserPassword)
	router.Get("/profile", h.profileUser)
	router.Get("/managers", h.getUsersManager)
	router.Route("/{userID}", func(router chi.Router) {
		router.Get("/", h.getUser)
		router.Put("/", h.updateUser)
		router.Patch("/update-role", h.updateRole)
		router.Delete("/", h.deleteUser)
		router.Post("/get-tasks", h.getAllTaskAssignedToUser)
	})
}

func (h *UserHandler) validateUserIDFromURLParam(r *http.Request) (int, error) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return 0, errors.New("user ID is required")
	}
	userID = strings.TrimLeft(userID, "0")
	userID = strings.Trim(userID, " ")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return 0, errors.New("cannot convert user ID from string to int, invalid user ID")
	}
	// Define a regular expression pattern to match the user ID format
	pattern := "^[0-9]+$"

	// Compile the regular expression pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return 0, err
	}

	// Check if the user ID matches the regular expression pattern
	if !regex.MatchString(userID) {
		return 0, errors.New("invalid user ID")
	}
	return id, nil

}
func (h *UserHandler) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserController.GetAllUsers(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)

}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.validateUserIDFromURLParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.UserController.GetUserByID(userID, ctx)
	if err != nil {
		if err == repository.ErrNoMatch {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	jsonBytes, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err := h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	userID, err := h.validateUserIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	err = h.UserController.DeleteUser(userID, ctx, r, tokenAuth, token)

	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	s := success{
		Status: "success",
	}
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	w.Write(jsonBytes)
}

func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.validateUserIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	userData := models.User{}

	// Read request body into a []byte variable
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse JSON request body into a User struct
	err = json.Unmarshal(body, &userData)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	user, err := h.UserController.UpdateUser(userID, userData, ctx)
	if err != nil {
		if err == repository.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func (h *UserHandler) updateRole(w http.ResponseWriter, r *http.Request) {
	err := h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	role := r.URL.Query().Get("role")
	if role == "" {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid role")))
		return
	}

	if role != "manager" && role != "user" {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("the role must be either 'manager' or 'user'")))
		return
	}
	userID, err := h.validateUserIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	user, err := h.UserController.UpdateRole(userID, role, ctx, r, tokenAuth)
	if err != nil {
		if err == repository.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func (h *UserHandler) changeUserPassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
	}
	oldPassword := r.PostForm.Get("oldPassword")
	newPassword := r.PostForm.Get("newPassword")
	token := GetToken(r, tokenAuth)

	email, ok := token.Get("email")
	if !ok {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("cannot get the email from token")))
	}
	// Convert email from interface{} to string
	emailStr, ok := email.(string)
	if !ok {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("cannot convert email from interface to string")))
		return
	}
	err = h.UserController.ChangeUserPassword(oldPassword, newPassword, emailStr, ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	s := success{
		Status: "success",
	}
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	w.Write(jsonBytes)
}

func (h *UserHandler) signup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
	}
	user := models.User{}
	user.Name = r.PostForm.Get("name")
	user.Email = r.PostForm.Get("email")
	user.Password = r.PostForm.Get("password")
	user.Role = "user"

	if user.Name == "" {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("missing name")))
		return
	}

	// Validate email
	if !h.isValidEmail(user.Email) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your email is not valid, please provide a valid email")))
		return
	}
	// Validate password
	// The password must contain at least 6 characters
	if !h.isValidPassword(user.Password) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your password is not valid, please provide a password that contains at least 6 characters")))
		return
	}

	token := MakeToken(user.Email, user.Password)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// Uncomment below for HTTPS:
		// Secure: true,
		Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
		Value: token,
	})
	r = r.WithContext(ctx)
	if err := h.UserController.AddUser(&user, ctx); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	_, err = json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte("Sign up successful"))
}

func (h *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data.", http.StatusInternalServerError)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	// Validate email
	if !h.isValidEmail(email) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your email is not valid, please provide a valid email")))
		return
	}
	// Validate password
	// The password must contain at least 6 characters
	if !h.isValidPassword(password) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your password is not valid, please provide a valid password that contains at least 6 characters")))
		return
	}
	// Generate a JWT token using the email and password
	token := MakeToken(email, password)

	// Check if the email and password are valid
	ok, err := h.UserController.CompareEmailAndPassword(email, password, ctx)
	if !ok {
		render.Render(w, r, ErrorRenderer(err))
		return
	}

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
}

func (h *UserHandler) logout(w http.ResponseWriter, r *http.Request) {
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
}

func (h *UserHandler) profileUser(w http.ResponseWriter, r *http.Request) {
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	userEmail, _ := token.Get("email")

	user, err := h.UserController.GetUserByEmail(userEmail.(string), ctx)
	if err != nil {
		if err == repository.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)

}

// Validates that an email address is in a valid format
func (h *UserHandler) isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Define a regular expression for validating email addresses
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Use the MatchString method to check if the email address matches the regular expression
	return emailRegex.MatchString(email)
}

// Validates that a password meets the minimum requirements
func (h *UserHandler) isValidPassword(password string) bool {
	if password == "" {
		return false
	}
	// Check if the password is at least 6 characters long
	return len(password) >= 6
}

func (h *UserHandler) getUsersManager(w http.ResponseWriter, r *http.Request) {
	err := h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	users, err := h.UserController.GetUsersManager(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
