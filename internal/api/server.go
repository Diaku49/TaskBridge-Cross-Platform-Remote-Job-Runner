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

	// Jobs
	mux.HandleFunc("POST /jobs", s.CreateJob)                      // create // test
	mux.HandleFunc("GET /jobs", s.ListJobs)                        // list // test
	mux.HandleFunc("GET /jobs/{jobId}", s.GetJob)                  // fetch one // test
	mux.HandleFunc("POST /jobs/{jobId}/result", s.SubmitJobResult) // submit job results

	// Agents
	mux.HandleFunc("POST /agents/register", s.RegisterAgent)            // register
	mux.HandleFunc("PUT /agents/{agentId}/heartbeat", s.AgentHeartbeat) // heartbeat
	mux.HandleFunc("GET /agents/{agentId}/next-job", s.AssignNextJob)   // assign next job
	mux.HandleFunc("GET /agents", s.ListAgents)                         // list agents

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"taskbridge-server"}`))
	})

	return mux
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
