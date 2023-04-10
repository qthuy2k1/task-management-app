package handler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"
	"github.com/qthuy2k1/task-management-app/models"
)

var taskCategoryIDKey = "taskCategoryID"

func taskCategories(router chi.Router) {
	router.Get("/", getAllTaskCategories)
	router.Post("/", createTaskCategory)
	router.Post("/csv", importTaskCategoryCSV)
	router.Route("/{taskCategoryID}", func(router chi.Router) {
		router.Get("/", getTaskCategory)
		router.Put("/", updateTaskCategory)
		router.Delete("/", deleteTaskCategory)
	})
}
func validateTaskCategoryIDFromURLParam(r *http.Request) (int, error) {
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

func createTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategory := &models.TaskCategory{}
	if err := render.Bind(r, taskCategory); err != nil {
		render.Render(w, r, ErrBadRequest)
		return
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

	if err := dbInstance.AddTaskCategory(taskCategory, r, tokenAuth, token); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, taskCategory); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}

func getAllTaskCategories(w http.ResponseWriter, r *http.Request) {
	taskCategories, err := dbInstance.GetAllTaskCategories(r, tokenAuth)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, taskCategories); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	taskCategory, err := dbInstance.GetTaskCategoryByID(taskCategoryID, r, tokenAuth)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	if err := render.Render(w, r, &taskCategory); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}

func deleteTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err = dbInstance.DeleteTaskCategory(taskCategoryID, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
}
func updateTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	taskCategoryData := models.TaskCategory{}
	if err := render.Bind(r, &taskCategoryData); err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	taskCategory, err := dbInstance.UpdateTaskCategory(taskCategoryID, taskCategoryData, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
	if err := render.Render(w, r, &taskCategory); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}

func importTaskCategoryCSV(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
	}
	path := r.PostForm.Get("path")
	taskCategoryList, err := dbInstance.ImportTaskCategoryDataFromCSV(path)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)

	for _, taskCategory := range taskCategoryList.TaskCategories {
		if err := dbInstance.AddTaskCategory(&taskCategory, r, tokenAuth, token); err != nil {
			render.Render(w, r, ErrorRenderer(err))
			return
		}
	}
	if err := render.Render(w, r, &taskCategoryList); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
}
