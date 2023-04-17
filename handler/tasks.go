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

func tasks(router chi.Router) {
	router.Get("/", getAllTasks)
	router.Post("/", createTask)
	// router.Post("/csv", importTaskCSV)
	router.Route("/{taskID}", func(router chi.Router) {
		router.Get("/", getTask)
		router.Put("/", updateTask)
		// router.Patch("/lock", lockTask)
		router.Delete("/", deleteTask)
		router.Put("/lock", lockTask)
		router.Put("/unlock", unLockTask)
		router.Post("/add-user", createUserTaskDetail)
		router.Post("/delete-user", deleteUserFromTask)
		router.Get("/get-users", getAllUserAsignnedToTask)
	})
}

func validateTaskIDFromURLParam(r *http.Request) (int, error) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		return 0, errors.New("task ID is required")
	}
	taskID = strings.TrimLeft(taskID, "0")
	taskID = strings.Trim(taskID, " ")
	id, err := strconv.Atoi(taskID)
	if err != nil {
		return 0, errors.New("cannot convert task ID from string to int, invalid task ID")
	}
	// Define a regular expression pattern to match the user ID format
	pattern := "^[0-9]+$"

	// Compile the regular expression pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return 0, err
	}

	// Check if the user ID matches the regular expression pattern
	if !regex.MatchString(taskID) {
		return 0, errors.New("invalid task ID")
	}
	return id, nil

}

func createTask(w http.ResponseWriter, r *http.Request) {
	task := models.Task{}
	// Read request body into a []byte variable
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Parse JSON request body into a User struct
	err = json.Unmarshal(body, &task)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	if err := dbInstance.AddTask(&task, ctx, r, tokenAuth, token); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	jsonBytes, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func getAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := dbInstance.GetAllTasks(ctx, r, tokenAuth)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonBytes, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := validateTaskIDFromURLParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := dbInstance.GetTaskByID(taskID, ctx, r, tokenAuth)
	if err != nil {
		if err == db.ErrNoMatch {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonBytes, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err = dbInstance.DeleteTask(taskID, ctx, r, tokenAuth, token)
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

func updateTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}

	taskData := models.Task{}
	// Read request body into a []byte variable
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse JSON request body into a User struct
	err = json.Unmarshal(body, &taskData)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
	task, err := dbInstance.UpdateTask(taskID, taskData, ctx, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no rows afftected")))
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
}

func lockTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err = dbInstance.LockTask(taskID, ctx, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
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

func unLockTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err = dbInstance.UnLockTask(taskID, ctx, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
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

// func importTaskCSV(w http.ResponseWriter, r *http.Request) {
// 	err := r.ParseForm()
// 	if err != nil {
// 		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
// 	}
// 	path := r.PostForm.Get("path")
// 	taskList, err := dbInstance.ImportTaskDataFromCSV(path)
// 	if err != nil {
// 		render.Render(w, r, ErrorRenderer(err))
// 		return
// 	}
// 	token := GetToken(r, tokenAuth)
// 	for _, task := range taskList.Tasks {
// 		if err := dbInstance.AddTask(&task, r, tokenAuth, token); err != nil {
// 			render.Render(w, r, ErrorRenderer(err))
// 			return
// 		}
// 		if err = dbInstance.AddUserToTask(task.AuthorID, task.ID, r, tokenAuth, token); err != nil {
// 			render.Render(w, r, ErrorRenderer(err))
// 			return
// 		}
// 	}
// 	if err := render.Render(w, r, &taskList); err != nil {
// 		render.Render(w, r, ErrorRenderer(err))
// 		return
// 	}
// }
