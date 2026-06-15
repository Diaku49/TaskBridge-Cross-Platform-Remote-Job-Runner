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
		HTTPError(w, http.StatusBadRequest, "invalid request body")
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
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	HTTPResponse(w, http.StatusCreated, job)
}

func (s *server) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.st.ListJobs()
	if err != nil {
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	HTTPResponse(w, http.StatusOK, jobs)
}

func (s *server) GetJob(w http.ResponseWriter, r *http.Request) {
	jobId := r.PathValue("jobId")

	job, found, err := s.st.GetJob(jobId)
	if err != nil {
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !found {
		HTTPError(w, http.StatusNotFound, JobNotFoundError)
		return
	}

	HTTPResponse(w, http.StatusOK, job)
}

func (s *server) SubmitJobResult(w http.ResponseWriter, r *http.Request) {
	jobId := r.PathValue("jobsId")
	var req model.SubmitJobResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if err := s.st.CompleteJob(jobId, req.Status, req.Logs, req.Result, req.Error); err != nil {
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	HTTPResponse(w, http.StatusOK, map[string]any{
		"message": "job result submitted",
	})
}

func (s *server) CancelJob(w http.ResponseWriter, r *http.Request) {}
