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

type TaskHandler struct {
	TaskController           *controllers.TaskController
	UserController           *controllers.UserController
	UserTaskDetailController *controllers.UserTaskDetailController
}

func NewTaskHandler(database *repositories.Database) *TaskHandler {
	taskRepository := repositories.NewTaskRepository(database)
	taskController := controllers.NewTaskController(taskRepository)
	userRepository := repositories.NewUserRepository(database)
	userController := controllers.NewUserController(userRepository)
	userTaskDetailRepository := repositories.NewUserTaskDetailRepository(database)
	userTaskDetailController := controllers.NewUserTaskDetailController(userTaskDetailRepository)
	return &TaskHandler{TaskController: taskController, UserController: userController, UserTaskDetailController: userTaskDetailController}
}

func (h *TaskHandler) tasks(router chi.Router) {
	router.Get("/", h.getAllTasks)
	router.Post("/", h.addTask)
	router.Post("/csv", h.importTaskCSV)
	router.Get("/filter-name", h.getTasksByName)
	// router.Get("/filter", h.filterTasks)
	// router.Get("/count-filtered-status", h.countFilteredStatusTask)
	router.Route("/{taskID}", func(router chi.Router) {
		router.Get("/", h.getTask)
		router.Put("/", h.updateTask)
		router.Delete("/", h.deleteTask)
		router.Patch("/lock", h.lockTask)
		router.Patch("/unlock", h.unLockTask)
		router.Post("/add-user", h.addUserTaskDetail)
		router.Post("/delete-user", h.deleteUserFromTask)
		router.Get("/get-users", h.getAllUserAsignnedToTask)
		router.Get("/get-task-category", h.getTaskCategoryOfTask)
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
	utils.RenderJson(w, task)
}

func (h *TaskHandler) getAllTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	queryParams := make(map[string]any)
	// set default for query
	queryParams["page"] = 1
	queryParams["size"] = 2
	queryParams["field"] = "id"
	queryParams["order"] = "asc"
	for key, values := range query {
		if len(values) > 0 {
			switch key {
			case "page":
				pageNumber, err := strconv.Atoi(values[0])
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				queryParams[key] = pageNumber
			case "size":
				pageSize, err := strconv.Atoi(values[0])
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				queryParams[key] = pageSize
			case "author_id":
				authorID, err := strconv.Atoi(values[0])
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				queryParams[key] = authorID
			default:
				queryParams[key] = values[0]
			}
		}
	}

	tasks, err := h.TaskController.GetAllTasks(ctx, queryParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RenderJson(w, tasks)
}

func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.validateTaskIDFromURLParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := h.TaskController.GetTaskByID(taskID, ctx)
	if err != nil {
		if err == repositories.ErrNoMatch {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	utils.RenderJson(w, task)

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
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("no rows afftected")))
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}

	utils.RenderJson(w, task)
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
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	s := success{
		Status: "success",
	}
	utils.RenderJson(w, s)
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
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	s := success{
		Status: "success",
	}
	utils.RenderJson(w, s)
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

	utils.RenderJson(w, taskList)
}

func (h *TaskHandler) getTaskCategoryOfTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.validateTaskIDFromURLParam(r)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	taskCategory, err := h.TaskController.GetTaskCategoryOfTask(taskID, ctx)
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

func (h *TaskHandler) getTasksByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid name")))
		return
	}
	tasks, err := h.TaskController.GetTasksByName(name, ctx)
	if err != nil {
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	utils.RenderJson(w, tasks)
}

func (h *TaskHandler) countFilteredStatusTask(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid status")))
		return
	}
	count, err := h.TaskController.CountFilteredStatusTask(status, ctx)
	if err != nil {
		if err == repositories.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	type CountResponse struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	response := CountResponse{
		Status: status,
		Count:  count,
	}
	utils.RenderJson(w, response)
}

// func (h *TaskHandler) filterTasks(w http.ResponseWriter, r *http.Request) {
// 	query := r.URL.Query()
// 	queryParams := make(map[string]string)
// 	for key, values := range query {
// 		if len(values) > 0 {
// 			queryParams[key] = values[0]
// 		}
// 	}
// 	tasks, err := h.TaskController.FilterTasks(queryParams, ctx)
// 	if err != nil {
// 		render.Render(w, r, ErrorRenderer(err))
// 		return
// 	}
// 	utils.RenderJson(w, tasks)
// }
