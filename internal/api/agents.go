package api

import (
	"encoding/json"
	"net/http"
	"taskbridge/internal/model"
	"time"
)

var LastSeenThreshold = 30 * time.Second

func (s *server) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	agent := model.Agent{
		ID:           req.ID,
		Hostname:     req.Hostname,
		OS:           req.OS,
		Arch:         req.Arch,
		Version:      req.Version,
		Capabilities: req.Capabilities,
	}

	agent, err := s.st.RegisterAgent(agent)
	if err != nil {
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	HTTPResponse(w, http.StatusCreated, agent)
}

func (s *server) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.st.ListAgents()
	if err != nil {
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	HTTPResponse(w, http.StatusOK, agents)
}

func (s *server) AssignNextJob(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentId")
	job, exist, err := s.st.AssignNextJob(agentID)
	if err != nil {
		HTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		HTTPError(w, http.StatusNoContent, "Assignable job not found")
		return
	}

	HTTPResponse(w, http.StatusAccepted, job)
}

func (s *server) AgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentId")
	if err := s.st.Heartbeat(agentID); err != nil {
		HTTPError(w, http.StatusNotFound, err.Error())
		return
	}

	HTTPResponse(w, http.StatusOK, nil)
}
