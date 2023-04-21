package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/internal/repositories"
	"github.com/qthuy2k1/task-management-app/internal/utils"
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

func (h *TaskHandler) addUserTaskDetail(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
		return
	}
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

	if err = h.UserTaskDetailController.AddUserToTask(userID, taskID, ctx); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	s := success{
		Status: "success",
	}
	utils.RenderJson(w, s)
}
func (h *TaskHandler) deleteUserFromTask(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("failed to parse form data")))
		return
	}
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
	token := GetToken(r, tokenAuth)
	if token == nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("no token found")))
		return
	}

	err = h.UserTaskDetailController.DeleteUserFromTask(userID, taskID, ctx)
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

func (h *TaskHandler) getAllUserAsignnedToTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid task id")))
		return
	}
	users, err := h.UserTaskDetailController.GetAllUsersAssignedToTask(taskID)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	utils.RenderJson(w, users)
}

func (h *UserHandler) getAllTaskAssignedToUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid user id")))
		return
	}
	tasks, err := h.UserTaskDetailController.GetAllTaskAssignedToUser(userID)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	utils.RenderJson(w, tasks)
}
