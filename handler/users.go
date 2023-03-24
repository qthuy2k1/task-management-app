package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"
	"github.com/qthuy2k1/task-management-app/models"
)

var userIDKey = "user_id"

func users(router chi.Router) {
	router.Get("/", getAllUsers)
	router.Post("/change-password", changeUserPassword)
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

func createUser(w http.ResponseWriter, r *http.Request, userData models.User) {
	if err := dbInstance.AddUser(&userData); err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, &userData); err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	http.Redirect(w, r, "/profile", http.StatusFound)
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := dbInstance.GetAllUsers(r, tokenAuth)
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
	err := dbInstance.DeleteUser(userId, r, tokenAuth)
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

func changeUserPassword(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	oldPassword := r.PostForm.Get("oldPassword")
	newPassword := r.PostForm.Get("newPassword")
	err := dbInstance.ChangeUserPassword(oldPassword, newPassword, r, tokenAuth)
	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
		return
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()
	user := models.User{}
	user.Name = r.PostForm.Get("name")
	user.Email = r.PostForm.Get("email")
	user.Password = r.PostForm.Get("password")
	user.Role = "user"
	context.WithValue(ctx, "user", user)

	if user.Email == "" || user.Name == "" || user.Password == "" {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("missing name, email or password")))
		return
	}

	// Validate email
	if !db.IsValidEmail(user.Email) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your email is not valid, please provide a valid email")))
		return
	}
	// Validate password
	// The password must contain at least 6 characters
	if !db.IsValidPassword(user.Password) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your password is not valid, please provide a password that contains at least 6 characters")))
		return
	}

	token := MakeToken(user.Email, user.Password)
	w.Write([]byte(token))

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// Uncomment below for HTTPS:
		// Secure: true,
		Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
		Value: token,
	})
	r = r.WithContext(ctx)
	createUser(w, r, user)

	w.Write([]byte("Sign up successful"))
	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)

}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data.", http.StatusInternalServerError)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	if email == "" || password == "" {
		http.Error(w, "Missing email or password.", http.StatusBadRequest)
		return
	}

	// Validate email
	if !db.IsValidEmail(email) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your email is not valid, please provide a valid email")))
		return
	}
	// Validate password
	// The password must contain at least 6 characters
	if !db.IsValidPassword(password) {
		render.Render(w, r, ErrorRenderer(fmt.Errorf("your password is not valid, please provide a valid password that contains at least 6 characters")))
		return
	}
	// Generate a JWT token using the email and password
	token := MakeToken(email, password)

	// Check if the email and password are valid
	ok := dbInstance.CompareEmailAndPassword(email, password, r, tokenAuth)
	if !ok {
		http.Error(w, "\nWrong email or password", http.StatusBadRequest)
		return
	}

	// Set the JWT token as a cookie in the response
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// Uncomment below for HTTPS:
		// Secure: true,
		Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
		Value: token,
	})
	// Write the response with a success message
	w.Write([]byte(fmt.Sprintf(`Login successful, your email is: %s and your token is: %s`, email, token)))

	// Redirect the user to the main page
	// http.Redirect(w, r, "/users/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		MaxAge:   -1, // Delete the cookie.
		SameSite: http.SameSiteLaxMode,
		// Uncomment below for HTTPS:
		// Secure: true,
		Name:  "jwt",
		Value: "",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
