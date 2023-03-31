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

var taskCategoryIDKey = "taskCategoryID"

func taskCategories(router chi.Router) {
	router.Get("/", getAllTaskCategories)
	router.Post("/", createTaskCategory)
	router.Route("/{taskCategoryID}", func(router chi.Router) {
		router.Use(TaskCategoryContext)
		router.Get("/", getTaskCategory)
		router.Put("/", updateTaskCategory)
		router.Delete("/", deleteTaskCategory)
	})
}
func TaskCategoryContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskCategoryID := chi.URLParam(r, "taskCategoryID")
		if taskCategoryID == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("task category ID is required")))
			return
		}
		id, err := strconv.Atoi(taskCategoryID)
		if err != nil {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task category ID")))
		}
		ctx := context.WithValue(r.Context(), taskCategoryIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createTaskCategory(w http.ResponseWriter, r *http.Request) {
	taskCategory := &models.TaskCategory{}
	if err := render.Bind(r, taskCategory); err != nil {
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
	taskCategoryID := r.Context().Value(taskCategoryIDKey).(int)
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
	taskCategoryID := r.Context().Value(taskCategoryIDKey).(int)
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}
	err := dbInstance.DeleteTaskCategory(taskCategoryID, r, tokenAuth, token)
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
	taskCategoryID := r.Context().Value(taskCategoryIDKey).(int)
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
