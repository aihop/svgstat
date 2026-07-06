package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/svgstat/svgstat/internal/auth"
	"github.com/svgstat/svgstat/internal/project"
	"github.com/svgstat/svgstat/internal/renderer"
)

func (a *App) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func (a *App) jsonSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (a *App) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
}

func (a *App) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func (a *App) handleSPA(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/spa.html")
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		a.jsonError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := a.auth.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if err == auth.ErrUserExists {
			a.jsonError(w, "User already exists", http.StatusConflict)
			return
		}
		log.Error().Err(err).Msg("Failed to register user")
		a.jsonError(w, "Failed to register", http.StatusInternalServerError)
		return
	}

	session, err := a.auth.CreateSession(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create session")
		a.jsonError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	a.setSessionCookie(w, session.Token)
	a.jsonSuccess(w, map[string]interface{}{
		"user":    user,
		"session": session,
	})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, err := a.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			a.jsonError(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Error().Err(err).Msg("Failed to login")
		a.jsonError(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	user, err := a.auth.GetUser(r.Context(), session.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		a.jsonError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	a.setSessionCookie(w, session.Token)
	a.jsonSuccess(w, map[string]interface{}{
		"user":    user,
		"session": session,
	})
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := a.getAuthToken(r)
	if token != "" {
		_ = a.auth.Logout(r.Context(), token)
	}
	a.clearSessionCookie(w)
	a.jsonSuccess(w, nil)
}

func (a *App) handleGetMe(w http.ResponseWriter, r *http.Request) {
	token := a.getAuthToken(r)
	if token == "" {
		a.jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.auth.ValidateSession(r.Context(), token)
	if err != nil {
		a.jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	a.jsonSuccess(w, user)
}

func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := a.getAuthToken(r)
		if token == "" {
			a.jsonError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := a.auth.ValidateSession(r.Context(), token)
		if err != nil {
			a.jsonError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) handleGetProjects(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.User)

	projects, err := a.projectRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list projects")
		a.jsonError(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	a.jsonSuccess(w, projects)
}

func (a *App) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.User)

	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Slug == "" {
		a.jsonError(w, "Name and slug are required", http.StatusBadRequest)
		return
	}

	b := make([]byte, 16)
	rand.Read(b)
	id := fmt.Sprintf("%x", b)

	// 生成唯一的 external_project_id
	extID := make([]byte, 16)
	rand.Read(extID)
	externalID := fmt.Sprintf("%x", extID)

	// 使用用户ID作为租户ID
	tenantID := user.ID

	p := &project.Project{
		ID:                id,
		UserID:            user.ID,
		ExternalProjectID: externalID,
		TenantID:          tenantID,
		Slug:              req.Slug,
		Name:              req.Name,
		Description:       req.Description,
		Status:            "active",
		Visibility:        "public",
		RenderEnabled:     true,
		BadgeEnabled:      true,
		WidgetEnabled:     true,
		ChartEnabled:      false,
	}

	if err := a.projectRepo.Create(r.Context(), p); err != nil {
		log.Error().Err(err).Msg("Failed to create project")
		a.jsonError(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	a.jsonSuccess(w, p)
}

func (a *App) handleGetProject(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.User)
	vars := mux.Vars(r)
	id := vars["id"]

	p, err := a.projectRepo.GetByIDAndUser(r.Context(), id, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get project")
		a.jsonError(w, "Failed to get project", http.StatusInternalServerError)
		return
	}

	if p == nil {
		a.jsonError(w, "Project not found", http.StatusNotFound)
		return
	}

	a.jsonSuccess(w, p)
}

func (a *App) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.User)
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	p, err := a.projectRepo.GetByIDAndUser(r.Context(), id, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get project")
		a.jsonError(w, "Failed to get project", http.StatusInternalServerError)
		return
	}

	if p == nil {
		a.jsonError(w, "Project not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Slug != "" {
		p.Slug = req.Slug
	}
	if req.Description != "" {
		p.Description = req.Description
	}

	if err := a.projectRepo.Update(r.Context(), p); err != nil {
		log.Error().Err(err).Msg("Failed to update project")
		a.jsonError(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	a.jsonSuccess(w, p)
}

func (a *App) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.User)
	vars := mux.Vars(r)
	id := vars["id"]

	if err := a.projectRepo.Delete(r.Context(), id, user.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete project")
		a.jsonError(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	a.jsonSuccess(w, nil)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (a *App) handleCounterSVG(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectSlug := vars["projectSlug"]
	name := vars["name"]

	proj, err := a.projectRepo.GetBySlug(r.Context(), projectSlug)
	if err != nil {
		log.Error().Err(err).Str("slug", projectSlug).Msg("Failed to get project")
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	if proj == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if !proj.RenderEnabled {
		http.Error(w, "Rendering disabled", http.StatusForbidden)
		return
	}

	value, err := a.counter.IncrementWithAnalytics(r.Context(), proj.ID, name)
	if err != nil {
		log.Error().Err(err).Str("project_id", proj.ID).Str("counter", name).Msg("Failed to increment counter")
		value = 0
	}

	_ = a.analytics.TrackRequest(r.Context(), r, proj.ID)

	color := r.URL.Query().Get("color")
	label := r.URL.Query().Get("label")

	svg, err := a.renderer.RenderCounter(renderer.CounterData{
		Value: value,
		Label: label,
		Color: color,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to render counter")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(svg))
}

func (a *App) handleBadgeSVG(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectSlug := vars["projectSlug"]
	name := vars["name"]

	proj, err := a.projectRepo.GetBySlug(r.Context(), projectSlug)
	if err != nil {
		log.Error().Err(err).Str("slug", projectSlug).Msg("Failed to get project")
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	if proj == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if !proj.BadgeEnabled {
		http.Error(w, "Badges disabled", http.StatusForbidden)
		return
	}

	value, err := a.counter.Increment(r.Context(), proj.ID, name)
	if err != nil {
		log.Error().Err(err).Str("project_id", proj.ID).Str("counter", name).Msg("Failed to increment counter")
		value = 0
	}

	_ = a.analytics.TrackRequest(r.Context(), r, proj.ID)

	color := r.URL.Query().Get("color")
	style := r.URL.Query().Get("style")
	label := r.URL.Query().Get("label")
	if label == "" {
		label = name
	}

	svg, err := a.renderer.RenderBadge(renderer.BadgeData{
		Label: label,
		Value: fmt.Sprintf("%d", value),
		Color: color,
		Style: style,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to render badge")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(svg))
}

func (a *App) handleGetProjectStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.User)
	vars := mux.Vars(r)
	id := vars["id"]

	p, err := a.projectRepo.GetByIDAndUser(r.Context(), id, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get project")
		a.jsonError(w, "Failed to get project", http.StatusInternalServerError)
		return
	}

	if p == nil {
		a.jsonError(w, "Project not found", http.StatusNotFound)
		return
	}

	stats, err := a.analytics.GetTodayStats(r.Context(), id)
	if err != nil {
		log.Error().Err(err).Str("project_id", id).Msg("Failed to get stats")
		a.jsonError(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	a.jsonSuccess(w, stats)
}
