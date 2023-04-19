package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/qthuy2k1/task-management-app/internal/repository"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

var tokenAuth *jwtauth.JWTAuth

const Secret = "<my-secret-key-1010>"

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(Secret), nil)
}

var ctx = context.Background()

func NewHandler(db *repository.Database) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.MethodNotAllowed(methodNotAllowedHandler)
	r.NotFound(notFoundHandler)
	userHandler := NewUserHandler(db)
	taskHandler := NewTaskHandler(db)
	taskCategoryHandler := NewTaskCategoryHandler(db)
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
		r.Route("/users", userHandler.users)
		r.Route("/task-categories", taskCategoryHandler.taskCategories)
		r.Route("/tasks", taskHandler.tasks)
	})

	// public routes
	r.Group(func(r chi.Router) {
		r.Get("/", welcome)
		r.Post("/signup", userHandler.signup)
		r.Post("/login", userHandler.login)
		r.Post("/logout", userHandler.logout)
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

func GetToken(r *http.Request, tokenAuth *jwtauth.JWTAuth) jwt.Token {
	token, err := tokenAuth.Decode(jwtauth.TokenFromCookie(r))
	if err != nil {
		return nil
	}
	return token
}

func welcome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome anonymous"))
}
