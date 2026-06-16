package api

import (
	"net/http"
	"taskbridge/internal/model"
	"time"
)

var LastSeenThreshold = 30 * time.Second

func (s *server) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterAgentRequest
	if errMsg := decodeAndValidateRequest(r, &req); errMsg != "" {
		HTTPError(w, http.StatusBadRequest, errMsg)
		return
	}
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
