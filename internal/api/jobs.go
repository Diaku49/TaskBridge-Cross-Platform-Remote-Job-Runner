package api

import (
	"encoding/json"
	"net/http"
	"taskbridge/internal/model"
)

var (
	JobNotFoundError = "job not found"
)

func (s *server) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req model.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	job := model.Job{
		Name:           req.Name,
		Type:           req.Type,
		Payload:        req.Payload,
		MaxRetries:     req.MaxRetries,
		TimeoutSeconds: req.TimeoutSeconds,
	}

	job, err := s.st.CreateJob(job)
	if err != nil {
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func (s *server) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.st.ListJobs()
	if err != nil {
		http.Error(w, "failed to list jobs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobs)
}

func (s *server) GetJob(w http.ResponseWriter, r *http.Request) {
	jobId := r.PathValue("jobId")

	job, found, err := s.st.GetJob(jobId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, JobNotFoundError, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

func (s *server) CancelJob(w http.ResponseWriter, r *http.Request) {}

func (s *server) SubmitJobResult(w http.ResponseWriter, r *http.Request) {}
