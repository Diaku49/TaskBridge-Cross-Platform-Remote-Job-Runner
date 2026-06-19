package api

import (
	"encoding/json"
	"log"
	"net/http"
	"taskbridge/internal/model"
	"taskbridge/internal/store"
)

type server struct {
	st store.Store
}

func NewServer(st store.Store) *server {
	return &server{
		st: st,
	}
}

func (s *server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /jobs", s.CreateJob)
	mux.HandleFunc("GET /jobs", s.ListJobs)
	mux.HandleFunc("GET /jobs/{jobId}", s.GetJob)
	mux.HandleFunc("POST /jobs/{jobId}/result", s.SubmitJobResult)

	mux.HandleFunc("POST /agents/register", s.RegisterAgent)
	mux.HandleFunc("POST /agents/{agentId}/heartbeat", s.AgentHeartbeat)
	mux.HandleFunc("POST /agents/{agentId}/next-job", s.AssignNextJob)
	mux.HandleFunc("GET /agents", s.ListAgents)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"taskbridge-server"}`))
	})

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HTTPError(w http.ResponseWriter, statusCode int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(model.ErrorResponse{
		Message: errMsg,
	}); err != nil {
		log.Printf("failed to write error JSON: %v", err)
	}
}

func HTTPResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}
