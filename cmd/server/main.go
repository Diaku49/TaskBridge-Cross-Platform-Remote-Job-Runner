package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"taskbridge/internal/api"
	"taskbridge/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "server listen address")
	flag.Parse()

	store := store.NewMemoryStore()
	server := api.NewServer(store)

	// TODO: Candidate should move route registration into internal/api.
	// Required routes:
	//   GET  /health
	//   POST /jobs
	//   GET  /jobs
	//   GET  /jobs/{jobId}
	//   POST /jobs/{jobId}/cancel
	//   POST /agents/register
	//   POST /agents/{agentId}/heartbeat
	//   POST /agents/{agentId}/next-job
	//   POST /jobs/{jobId}/result

	fmt.Printf("TaskBridge server listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Routes()))
}
