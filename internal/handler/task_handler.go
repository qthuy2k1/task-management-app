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
	"github.com/qthuy2k1/task-management-app/internal/controller"
	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repository"
)

type TaskHandler struct {
	TaskController           *controller.TaskController
	UserController           *controller.UserController
	UserTaskDetailController *controller.UserTaskDetailController
}

func NewTaskHandler(database *repository.Database) *TaskHandler {
	taskRepository := repository.NewTaskRepository(database)
	taskController := controller.NewTaskController(taskRepository)
	userRepository := repository.NewUserRepository(database)
	userController := controller.NewUserController(userRepository)
	userTaskDetailRepository := repository.NewUserTaskDetailRepository(database)
	userTaskDetailController := controller.NewUserTaskDetailController(userTaskDetailRepository)
	return &TaskHandler{TaskController: taskController, UserController: userController, UserTaskDetailController: userTaskDetailController}
}

func (h *TaskHandler) tasks(router chi.Router) {
	router.Get("/", h.getAllTasks)
	router.Post("/", h.addTask)
	router.Post("/csv", h.importTaskCSV)
	router.Route("/{taskID}", func(router chi.Router) {
		router.Get("/", h.getTask)
		router.Put("/", h.updateTask)
		router.Delete("/", h.deleteTask)
		router.Patch("/lock", h.lockTask)
		router.Patch("/unlock", h.unLockTask)
		router.Post("/add-user", h.addUserTaskDetail)
		router.Post("/delete-user", h.deleteUserFromTask)
		router.Get("/get-users", h.getAllUserAsignnedToTask)
	})
}

func (h *TaskHandler) validateTaskIDFromURLParam(r *http.Request) (int, error) {
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

func (h *TaskHandler) addTask(w http.ResponseWriter, r *http.Request) {
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
	err = h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	if err := h.TaskController.AddTask(&task, ctx); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	// Also add author to this task
	if err = h.UserTaskDetailController.AddUserToTask(task.AuthorID, task.ID, ctx); err != nil {
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

func (h *TaskHandler) getAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.TaskController.GetAllTasks(ctx)

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

func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.validateTaskIDFromURLParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := h.TaskController.GetTaskByID(taskID, ctx)
	if err != nil {
		if err == repository.ErrNoMatch {
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

func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
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
	err = h.TaskController.DeleteTask(taskID, ctx)
	if err != nil {
		if err == repository.ErrNoMatch {
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

func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.validateTaskIDFromURLParam(r)
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
	var isManager bool
	err = h.UserController.IsManager(ctx, r, tokenAuth)
	if err != nil {
		isManager = false
	} else {
		isManager = true
	}
	task, err := h.TaskController.UpdateTask(taskID, taskData, ctx, isManager)
	if err != nil {
		if err == repository.ErrNoMatch {
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

func (h *TaskHandler) lockTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
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
	err = h.TaskController.LockTask(taskID, ctx)
	if err != nil {
		if err == repository.ErrNoMatch {
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

func (h *TaskHandler) unLockTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
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
	}
	err = h.TaskController.UnLockTask(taskID, ctx)
	if err != nil {
		if err == repository.ErrNoMatch {
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

func (h *TaskHandler) importTaskCSV(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
	}
	path := r.PostForm.Get("path")
	taskList, err := h.TaskController.ImportTaskDataFromCSV(path)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
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
	}
	for _, task := range taskList {
		if err := h.TaskController.AddTask(&task, ctx); err != nil {
			render.Render(w, r, ErrorRenderer(err))
			return
		}
		if err = h.UserTaskDetailController.AddUserToTask(task.AuthorID, task.ID, ctx); err != nil {
			render.Render(w, r, ErrorRenderer(err))
			return
		}
	}

	jsonBytes, err := json.Marshal(taskList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
