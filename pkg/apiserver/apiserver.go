package apiserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"

	"superminikube/pkg/apiserver/pod"
	"superminikube/pkg/apiserver/watch"
)

func loggingMiddleware(next http.Handler) http.Handler {
	// log some metadata about the request
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("received request",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) Shutdown() {
	slog.Info("shutting down apiserver")
	s.server.Close()
}

func Start() error {
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	s, err := NewAPIServer()
	go func() {
		<-sigCtx.Done()
		s.Shutdown()
	}()
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}
	slog.Info("starting API server...")
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	r.Use(loggingMiddleware)
	// TODO: This is proof that notify needs to exist elsewhere...
	watchService := watch.NewService()
	podService := pod.NewService(s.redisClient, watchService)
	podHandler := pod.NewHandler(podService)
	api.HandleFunc("/pod", podHandler.CreatePod).Queries("nodename", "{nodename}")
	api.HandleFunc("/pod", podHandler.GetPod).Queries("nodename", "{nodename}", "uid", "{uid}")
	// post is probably the better verb here
	api.HandleFunc("/watch", watchService.WatchHandler).Methods(http.MethodGet)
	// TODO: will eventually initialize server inside NewAPIServer()
	s.server = &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	slog.Info("server listening on :8080")
	err = s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("API server failed to start: %w", err)
	}
	return nil
}

func NewAPIServer() (*APIServer, error) {
	return &APIServer{
		redisClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // TODO: make configurable, got so many options to worry about now
		}),
	}, nil
}

type APIServer struct {
	server      *http.Server
	redisClient *redis.Client
	// TODO
	config interface{}
}
