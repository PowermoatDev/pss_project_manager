package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"printsec-warroom/backend/internal/config"
	"printsec-warroom/backend/internal/models"
	"printsec-warroom/backend/internal/repository"
)

type Server struct {
	repo *repository.Repository
	cfg  config.Config
	mux  *http.ServeMux
}

func New(repo *repository.Repository, cfg config.Config) *Server {
	s := &Server{repo: repo, cfg: cfg, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.withCORS(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/projects", s.listProjects)
	s.mux.HandleFunc("POST /api/projects", s.createProject)
	s.mux.HandleFunc("PUT /api/projects/{id}", s.updateProject)
	s.mux.HandleFunc("POST /api/projects/{id}/close", s.closeProject)
	s.mux.HandleFunc("POST /api/projects/{id}/files", s.uploadProjectFile)
	s.mux.HandleFunc("GET /api/events", s.listEvents)
	s.mux.HandleFunc("POST /api/events", s.createEvent)
	s.mux.HandleFunc("PUT /api/events/{id}", s.updateEvent)
	s.mux.HandleFunc("DELETE /api/events/{id}", s.deleteEvent)
	s.mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.cfg.UploadDir))))
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.cfg.AllowedOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "time": time.Now().Format(time.RFC3339)})
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.repo.ListProjects(r.Context(), r.URL.Query().Get("view"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var input models.ProjectInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	project, err := s.repo.CreateProject(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var input models.ProjectInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	project, err := s.repo.UpdateProject(r.Context(), id, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) closeProject(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var payload struct {
		InstallDate string `json:"installDate"`
		Editor      string `json:"editor"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, err)
		return
	}
	if payload.Editor == "" {
		payload.Editor = "KC"
	}
	project, err := s.repo.CloseProject(r.Context(), id, payload.InstallDate, payload.Editor)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) uploadProjectFile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	kind := r.URL.Query().Get("kind")
	if kind != "quote" && kind != "custom" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind must be quote or custom"})
		return
	}
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		writeError(w, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, err)
		return
	}
	defer file.Close()

	dir := filepath.Join(s.cfg.UploadDir, strconv.FormatInt(id, 10))
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, err)
		return
	}
	safeName := safeFileName(header.Filename)
	storedName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), safeName)
	target := filepath.Join(dir, storedName)
	out, err := os.Create(target)
	if err != nil {
		writeError(w, err)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		writeError(w, err)
		return
	}
	result, err := s.repo.SaveFile(r.Context(), id, kind, safeName, "/uploads/"+strconv.FormatInt(id, 10)+"/"+storedName, r.FormValue("uploadedBy"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.repo.ListEvents(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	var input models.CalendarEventInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	event, err := s.repo.CreateEvent(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var input models.CalendarEventInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	event, err := s.repo.UpdateEvent(r.Context(), id, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.repo.DeleteEvent(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func readJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, sql.ErrNoRows) {
		status = http.StatusNotFound
	} else if strings.Contains(strings.ToLower(err.Error()), "invalid") || strings.Contains(strings.ToLower(err.Error()), "bad request") {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func safeFileName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(name)
}
