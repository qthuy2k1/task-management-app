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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"
	models "github.com/qthuy2k1/task-management-app/models/gen"
)

func taskCategories(router chi.Router) {
	router.Get("/", getAllTaskCategories)
	router.Post("/", createTaskCategory)
	// router.Post("/csv", importTaskCategoryCSV)
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

	if err := dbInstance.AddTaskCategory(taskCategory, ctx, r, tokenAuth, token); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	jsonBytes, err := json.Marshal(taskCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func getAllTaskCategories(w http.ResponseWriter, r *http.Request) {
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	taskCategories, err := dbInstance.GetAllTaskCategories(ctx, r, tokenAuth, token)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	jsonBytes, err := json.Marshal(taskCategories)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func getTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	taskCategory, err := dbInstance.GetTaskCategoryByID(taskCategoryID, ctx, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	jsonBytes, err := json.Marshal(taskCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
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
	err = dbInstance.DeleteTaskCategory(taskCategoryID, ctx, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
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
func updateTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategoryID, err := validateTaskCategoryIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
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
	taskCategory, err := dbInstance.UpdateTaskCategory(taskCategoryID, taskCategoryData, ctx, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}

	jsonBytes, err := json.Marshal(taskCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)

}

// func importTaskCategoryCSV(w http.ResponseWriter, r *http.Request) {
// 	err := r.ParseForm()
// 	if err != nil {
// 		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
// 	}
// 	path := r.PostForm.Get("path")
// 	taskCategoryList, err := dbInstance.ImportTaskCategoryDataFromCSV(path)
// 	if err != nil {
// 		render.Render(w, r, ErrorRenderer(err))
// 	}
// 	token := GetToken(r, tokenAuth)

// 	for _, taskCategory := range taskCategoryList.TaskCategories {
// 		if err := dbInstance.AddTaskCategory(&taskCategory, r, tokenAuth, token); err != nil {
// 			render.Render(w, r, ErrorRenderer(err))
// 			return
// 		}
// 	}
// 	if err := render.Render(w, r, &taskCategoryList); err != nil {
// 		render.Render(w, r, ErrorRenderer(err))
// 		return
// 	}
// }
