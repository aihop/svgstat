package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *App) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.Use(a.loggingMiddleware)

	r.HandleFunc("/health", a.handleHealth).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.PathPrefix("/components/").Handler(http.StripPrefix("/components/", http.FileServer(http.Dir("web/components"))))

	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/auth/register", a.handleRegister).Methods("POST")
	api.HandleFunc("/auth/login", a.handleLogin).Methods("POST")
	api.HandleFunc("/auth/logout", a.handleLogout).Methods("POST")
	api.HandleFunc("/auth/me", a.handleGetMe).Methods("GET")

	projects := api.PathPrefix("/projects").Subrouter()
	projects.Use(a.authMiddleware)
	projects.HandleFunc("", a.handleGetProjects).Methods("GET")
	projects.HandleFunc("", a.handleCreateProject).Methods("POST")
	projects.HandleFunc("/{id}", a.handleGetProject).Methods("GET")
	projects.HandleFunc("/{id}", a.handleUpdateProject).Methods("PUT")
	projects.HandleFunc("/{id}", a.handleDeleteProject).Methods("DELETE")
	projects.HandleFunc("/{id}/stats", a.handleGetProjectStats).Methods("GET")

	r.HandleFunc("/svg/{projectSlug}/counter/{name}.svg", a.handleCounterSVG).Methods("GET")
	r.HandleFunc("/svg/{projectSlug}/badge/{name}.svg", a.handleBadgeSVG).Methods("GET")

	r.PathPrefix("/").HandlerFunc(a.handleSPA).Methods("GET")

	return r
}
