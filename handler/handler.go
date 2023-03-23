package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"
	"github.com/qthuy2k1/task-management-app/models"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

var dbInstance db.Database
var tokenAuth *jwtauth.JWTAuth

const Secret = "<my-secret-key-1010>"

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(Secret), nil)
}

func NewHandler(db db.Database) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	dbInstance = db
	r.MethodNotAllowed(methodNotAllowedHandler)
	r.NotFound(notFoundHandler)

	// protected routes
	r.Group(func(r chi.Router) {
		/*

			jwtauth.Verifier automatically searches for a JWT token in an incoming request:
			- The jwt URI query parameter
				- The Authorization: Bearer <token> request header
				- The jwt cookie

		*/

		r.Use(jwtauth.Verifier(tokenAuth))
		// send 401 Unauthorized response for any unverified
		r.Use(jwtauth.Authenticator)

		r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", strings.Split(claims["email"].(string), "@")[0])))
			println(claims["email"].(string))
		})
		r.Route("/users", users)
		r.Route("/task-categories", taskCategories)
		r.Route("/tasks", tasks)
	})

	// public routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome anonymous"))
		})
		r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
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

		})
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			r.ParseForm()
			user := models.User{}
			user.Email = r.PostForm.Get("email")
			context.WithValue(ctx, "email", user.Email)
			user.Password = r.PostForm.Get("password")

			if user.Email == "" || user.Password == "" {
				http.Error(w, "Missing email or password.", http.StatusBadRequest)
				return
			}

			token := MakeToken(user.Email, user.Password)

			ok := db.CompareEmailAndPassword(user.Email, user.Password, r, tokenAuth)
			if !ok {
				http.Error(w, "\nWrong email or password", http.StatusBadRequest)
				return
			}

			http.SetCookie(w, &http.Cookie{
				HttpOnly: true,
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				SameSite: http.SameSiteLaxMode,
				// Uncomment below for HTTPS:
				// Secure: true,
				Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
				Value: token,
			})
			w.Write([]byte(fmt.Sprintf(`Login successful, your email is: %s and your token is: %s`, user.Email, token)))

			// http.Redirect(w, r, "/users/", http.StatusSeeOther)
		})
		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
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
		})
	})
	return r
}
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(405)
	render.Render(w, r, ErrMethodNotAllowed)
}
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(400)
	render.Render(w, r, ErrNotFound)
}

func MakeToken(email, password string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{
		"email":    email,
		"password": password,
	})
	return tokenString
}
