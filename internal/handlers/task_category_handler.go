package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/internal/controllers"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repositories"
	"github.com/qthuy2k1/task-management-app/internal/utils"
)

type TaskCategoryHandler struct {
	TaskCategoryController *controllers.TaskCategoryController
	UserController         *controllers.UserController
}

func NewTaskCategoryHandler(database *repositories.Database) *TaskCategoryHandler {
	taskCategoryRepository := repositories.NewTaskCategoryRepository(database)
	taskCategoryController := controllers.NewTaskCategoryController(taskCategoryRepository)
	userRepository := repositories.NewUserRepository(database)
	userController := controllers.NewUserController(userRepository)
	return &TaskCategoryHandler{TaskCategoryController: taskCategoryController, UserController: userController}
}

func (h *TaskCategoryHandler) taskCategories(router chi.Router) {
	router.Get("/", h.getAllTaskCategories)
	router.Post("/", h.addTaskCategory)
	router.Post("/csv", h.importTaskCategoryCSV)
	router.Route("/{taskCategoryID}", func(router chi.Router) {
		router.Get("/", h.getTaskCategory)
		router.Put("/", h.updateTaskCategory)
		router.Delete("/", h.deleteTaskCategory)
	})
}
func (h *TaskCategoryHandler) validateTaskCategoryIDFromURLParam(r *http.Request) (int, error) {
	taskCategoryID := chi.URLParam(r, "taskCategoryID")
	if taskCategoryID == "" {
		return 0, errors.New("task category ID is required")
	}
	taskCategoryID = strings.TrimLeft(taskCategoryID, "0")
	taskCategoryID = strings.Trim(taskCategoryID, " ")
	id, err := strconv.Atoi(taskCategoryID)
	if err != nil {
		return 0, errors.New("cannot convert task category ID from string to int, invalid task category ID")
	}
	// Define a regular expression pattern to match the user ID format
	pattern := "^[0-9]+$"

	// Compile the regular expression pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return 0, err
	}

	// Check if the user ID matches the regular expression pattern
	if !regex.MatchString(taskCategoryID) {
		return 0, errors.New("invalid task category ID")
	}
	return id, nil

}

func (h *TaskCategoryHandler) addTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategory := &models.TaskCategory{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Parse JSON request body into a User struct
	err = json.Unmarshal(body, &taskCategory)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	if taskCategory.Name == "" {
		render.Render(w, r, ErrBadRequest)
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}

	err = h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}

	if err := h.TaskCategoryController.AddTaskCategory(taskCategory, ctx); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	utils.RenderJson(w, taskCategory)
}

func (h *TaskCategoryHandler) getAllTaskCategories(w http.ResponseWriter, r *http.Request) {
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}

	taskCategories, err := h.TaskCategoryController.GetAllTaskCategories(ctx)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	utils.RenderJson(w, taskCategories)
}

func (h *TaskCategoryHandler) getTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := h.validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	taskCategory, err := h.TaskCategoryController.GetTaskCategoryByID(taskCategoryID, ctx)
	if err != nil {
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	utils.RenderJson(w, taskCategory)
}

func (h *TaskCategoryHandler) deleteTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := h.validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err = h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	err = h.TaskCategoryController.DeleteTaskCategory(taskCategoryID, ctx)
	if err != nil {
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
	s := success{
		Status: "success",
	}
	utils.RenderJson(w, s)
}
func (h *TaskCategoryHandler) updateTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := h.validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err = h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	taskCategoryData := models.TaskCategory{}
	// Read request body into a []byte variable
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &taskCategoryData)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	taskCategory, err := h.TaskCategoryController.UpdateTaskCategory(taskCategoryID, taskCategoryData, ctx)
	if err != nil {
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}

	utils.RenderJson(w, taskCategory)
}

func (h *TaskCategoryHandler) importTaskCategoryCSV(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
	}
	path := r.PostForm.Get("path")
	taskCategoryList, err := h.TaskCategoryController.ImportTaskCategoryDataFromCSV(path)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err = h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	for _, taskCategory := range taskCategoryList {
		if err := h.TaskCategoryController.AddTaskCategory(&taskCategory, ctx); err != nil {
			render.Render(w, r, ErrorRenderer(err))
			return
		}
	}
	utils.RenderJson(w, taskCategoryList)
}
