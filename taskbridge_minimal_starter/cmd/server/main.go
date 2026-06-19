package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"taskbridge/internal/api"
	"taskbridge/internal/store"
	"taskbridge/internal/store/memory"
	"taskbridge/internal/store/sqlite"
)

func main() {
	addr := flag.String("addr", ":8080", "server listen address")
	storeKind := flag.String("store", envOrDefault("TASKBRIDGE_STORE", "memory"), "store backend: memory or sqlite")
	sqlitePath := flag.String("sqlite-path", envOrDefault("TASKBRIDGE_SQLITE_PATH", "taskbridge.db"), "sqlite database path")
	flag.Parse()

	st, closeStore, err := newStore(context.Background(), *storeKind, *sqlitePath)
	if err != nil {
		log.Fatal(err)
	}
	defer closeStore()

	server := api.NewServer(st)

	fmt.Printf("TaskBridge server listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Routes()))
}

func newStore(ctx context.Context, kind string, sqlitePath string) (store.Store, func(), error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "memory":
		return memory.NewMemoryStore(), func() {}, nil
	case "sqlite":
		st, err := sqlite.NewSqliteStore(ctx, sqlitePath)
		if err != nil {
			return nil, nil, err
		}
		return st, func() { _ = st.Close() }, nil
	default:
		return nil, nil, fmt.Errorf("unsupported store backend: %s", kind)
	}
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
