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
	s.watchService.Shutdown()
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
	// TODO: A cleaner pattern may be to setup service specific routers in their own packages
	// and 'add' them all to this route
	// may also want a root route like /api
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	r.Use(loggingMiddleware)
	api.HandleFunc("/pod", s.PodHandler).Queries("nodename", "{nodename}")
	api.HandleFunc("/pod", s.PodHandler).Queries("nodename", "{nodename}", "uid", "{uid}")
	// post is probably the better verb here
	api.HandleFunc("/watch", s.watchService.WatchHandler).Methods(http.MethodGet) // eventually will be watching per node
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
			Addr: "host.docker.internal:6379", // TODO: make configurable, got so many options to worry about now
		}),
		watchService: watch.New(),
	}, nil
}

type APIServer struct {
	server       *http.Server
	redisClient  *redis.Client
	watchService *watch.WatchService
}
