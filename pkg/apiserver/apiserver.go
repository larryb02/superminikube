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

// Setup configures routes and initializes the HTTP server.
func (s *APIServer) Setup() {
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

	s.server = &http.Server{
		Addr:    s.opts.Addr,
		Handler: r,
	}
}

// ListenAndServe starts the server. Blocks until server stops.
func (s *APIServer) ListenAndServe() error {
	slog.Info("server listening", slog.String("addr", s.opts.Addr))
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("API server failed: %w", err)
	}
	return nil
}

func Start() error {
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s, err := NewAPIServer(APIServerOpts{Addr: ":8080"})
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	go func() {
		<-sigCtx.Done()
		s.Shutdown()
	}()

	slog.Info("starting API server...")
	s.Setup()
	return s.ListenAndServe()
}

func NewAPIServer(opts APIServerOpts) (*APIServer, error) {
	return &APIServer{
		redisClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // TODO: make configurable, got so many options to worry about now
		}),
		opts: opts,
	}, nil
}


type APIServer struct {
	server      *http.Server
	redisClient *redis.Client
	opts        APIServerOpts
}

type APIServerOpts struct {
	Addr string
}
