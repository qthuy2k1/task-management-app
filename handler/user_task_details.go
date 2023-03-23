package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"
)

var userTaskDetailIDKey = "userTaskDetailID"

func UserTaskDetailContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userTaskDetailID := chi.URLParam(r, "userTaskDetailID")
		if userTaskDetailID == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("task category ID is required")))
			return
		}
		id, err := strconv.Atoi(userTaskDetailID)
		if err != nil {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task category ID")))
		}
		ctx := context.WithValue(r.Context(), userTaskDetailIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createUserTaskDetail(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userID, err := strconv.Atoi(r.PostForm.Get("id"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid user id")))
		return
	}
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task id")))
		return
	}
	if err = dbInstance.AddUserToTask(userID, taskID, r, tokenAuth); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
}
func deleteUserFromTask(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userID, err := strconv.Atoi(r.PostForm.Get("id"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid user id")))
		return
	}
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task id")))
		return
	}
	err = dbInstance.DeleteUserFromTask(userID, taskID, r, tokenAuth)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
}

func getAllUserAsignnedToTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task id")))
		return
	}
	users, err := dbInstance.GetAllUsersAssignedToTask(taskID)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, users); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getAllTaskAssignedToUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid user id")))
		return
	}
	tasks, err := dbInstance.GetAllTaskAssignedToUser(userID)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, tasks); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}
