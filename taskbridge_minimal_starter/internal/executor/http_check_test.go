package executor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func TestHTTPCheckExecutorExecuteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ex := executor.NewHTTPCheckExecutor(server.Client())
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobHTTPCheck,
		Payload: map[string]any{
			"url":             server.URL,
			"expected_status": http.StatusNoContent,
		},
	})

	if result.Status != model.JobSuccess {
		t.Fatalf("expected status %s, got %s: %s", model.JobSuccess, result.Status, result.Error)
	}
	if result.Result["received_status"] != http.StatusNoContent {
		t.Fatalf("expected received_status %d, got %v", http.StatusNoContent, result.Result["received_status"])
	}
}

func TestHTTPCheckExecutorExecuteUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ex := executor.NewHTTPCheckExecutor(server.Client())
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobHTTPCheck,
		Payload: map[string]any{
			"url":             server.URL,
			"expected_status": http.StatusOK,
		},
	})

	if result.Status != model.JobFailed {
		t.Fatalf("expected status %s, got %s", model.JobFailed, result.Status)
	}
	if result.Error != "unexpected status code" {
		t.Fatalf("expected unexpected status code error, got %q", result.Error)
	}
	if result.Result["received_status"] != http.StatusInternalServerError {
		t.Fatalf("expected received_status %d, got %v", http.StatusInternalServerError, result.Result["received_status"])
	}
}
