package api

import (
	"net/http"
	"time"
)

var LastSeenThreshold = 30 * time.Second

func (s *server) RegisterAgent(w http.ResponseWriter, r *http.Request) {}

func (s *server) ListAgents(w http.ResponseWriter, r *http.Request) {}

func (s *server) AssignNextJob(w http.ResponseWriter, r *http.Request) {}

func (s *server) AgentHeartbeat(w http.ResponseWriter, r *http.Request) {}
