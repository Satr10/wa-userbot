package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type HealthStatus struct {
	Status string `json:"status"`
}

var server *http.Server

func StartServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/healthcheck", healthCheckOK)

	server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	logrus.Infof("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}

func healthCheckOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func Shutdown(ctx context.Context) error {
	logrus.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{Status: "running"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		logrus.Errorf("Failed to write health check response: %v", err)
		http.Error(w, "Failed to write health check response", http.StatusInternalServerError)
	}
}
