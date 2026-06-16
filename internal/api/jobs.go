package api

import (
	"net/http"
	"taskbridge/internal/model"
)

var (
	JobNotFoundError = "job not found"
)

func (s *server) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req model.CreateJobRequest
	if errMsg := decodeAndValidateRequest(r, &req); errMsg != "" {
		HTTPError(w, http.StatusBadRequest, errMsg)
		return
	}

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
	jobId := r.PathValue("jobId")
	var req model.SubmitJobResultRequest
	if errMsg := decodeAndValidateRequest(r, &req); errMsg != "" {
		HTTPError(w, http.StatusBadRequest, errMsg)
		return
	}
	if err := s.st.CompleteJob(jobId, req.Status, req.Logs, req.Result, req.Error); err != nil {
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	HTTPResponse(w, http.StatusOK, map[string]any{
		"message": "job result submitted",
	})
}

func (s *server) CancelJob(w http.ResponseWriter, r *http.Request) {}
