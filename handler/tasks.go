package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"
	"github.com/qthuy2k1/task-management-app/models"
)

var taskIDKey = "taskID"

func tasks(router chi.Router) {
	router.Get("/", getAllTasks)
	router.Post("/", createTask)
	router.Route("/{taskID}", func(router chi.Router) {
		router.Use(TaskContext)
		router.Get("/", getTask)
		router.Put("/", updateTask)
		// router.Patch("/lock", lockTask)
		router.Delete("/", deleteTask)
		router.Put("/lock", lockTask)
		router.Put("/unlock", unLockTask)
		router.Post("/add-user", createUserTaskDetail)
		router.Post("/delete-user", deleteUserFromTask)
		router.Post("/get-users", getAllUserAsignnedToTask)
	})
}
func TaskContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID := chi.URLParam(r, "taskID")
		if taskID == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("task ID is required")))
			return
		}
		id, err := strconv.Atoi(taskID)
		if err != nil {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task ID")))
		}
		ctx := context.WithValue(r.Context(), taskIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createTask(w http.ResponseWriter, r *http.Request) {
	task := &models.Task{}
	if err := render.Bind(r, task); err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	if err := dbInstance.AddTask(task, r, tokenAuth, token); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, task); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}

func getAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := dbInstance.GetAllTasks(r, tokenAuth)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, tasks); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.Context().Value(taskIDKey).(int)
	task, err := dbInstance.GetTaskByID(taskID, r, tokenAuth)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	if err := render.Render(w, r, &task); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.Context().Value(taskIDKey).(int)
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err := dbInstance.DeleteTask(taskID, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
}
func updateTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.Context().Value(taskIDKey).(int)
	taskData := models.Task{}
	if err := render.Bind(r, &taskData); err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	task, err := dbInstance.UpdateTask(taskID, taskData, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("you are not the manager")))
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
	if err := render.Render(w, r, &task); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}

func lockTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.Context().Value(taskIDKey).(int)
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err := dbInstance.LockTask(taskID, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
}

func unLockTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.Context().Value(taskIDKey).(int)
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err := dbInstance.UnLockTask(taskID, r, tokenAuth, token)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
}
