package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"
	"github.com/qthuy2k1/task-management-app/models"
)

var userIDKey = "user_id"

func users(router chi.Router) {
	router.Get("/", getAllUsers)
	router.Post("/", createUser)
	router.Route("/{userID}", func(router chi.Router) {
		router.Use(UserContext)
		router.Get("/", getUser)
		router.Put("/", updateUser)
		router.Delete("/", deleteUser)
		router.Post("/get-tasks", getAllTaskAssignedToUser)
	})
}

func UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("user ID is required")))
			return
		}
		id, err := strconv.Atoi(userID)
		if err != nil {
			render.Render(w, r, ErrorRenderer(fmt.Errorf("invalid user ID")))
		}
		ctx := context.WithValue(r.Context(), userIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// user := r.Context().Value("user").(*models.User)
	// token, _, _ := jwtauth.FromContext(r.Context())
	user := &models.User{}
	user.Token = jwtauth.TokenFromCookie(r)
	if err := render.Bind(r, user); err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	if err := dbInstance.AddUser(user, r); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, user); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	http.Redirect(w, r, "/profile", http.StatusFound)
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := dbInstance.GetAllUsers()
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, users); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	user, err := dbInstance.GetUserByID(userID)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
		return
	}
	if err := render.Render(w, r, &user); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(userIDKey).(int)
	err := dbInstance.DeleteUser(userId, r)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
}
func updateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	print(userID)
	userData := models.User{}
	if err := render.Bind(r, &userData); err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}
	user, err := dbInstance.UpdateUser(userID, userData)
	if err != nil {
		if err == db.ErrNoMatch {
			render.Render(w, r, ErrNotFound)
		} else {
			render.Render(w, r, ServerErrorRenderer(err))
		}
		return
	}
	if err := render.Render(w, r, &user); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
}
