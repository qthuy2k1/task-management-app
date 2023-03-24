package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/qthuy2k1/task-management-app/db"

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
		r.Post("/signup", signup)
		r.Post("/login", login)
		r.Post("/logout", logout)
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
