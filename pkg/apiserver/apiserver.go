package apiserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"superminikube/pkg/apiserver/watch"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
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
	s.cancel()
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
	r.Use(loggingMiddleware)
	r.HandleFunc("/pod", s.PodHandler).Queries("nodename", "{nodename}")
	r.HandleFunc("/pod", s.PodHandler).Queries("nodename", "{nodename}", "uid", "{uid}")
	// post is probably the better verb here
	r.HandleFunc("/watch", s.watchService.WatchHandler).Methods(http.MethodGet) // eventually will be watching per node
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
	ctx, cancel := context.WithCancel(context.Background()) // should probably get context from entry point
	return &APIServer{
		ctx:    ctx,
		cancel: cancel,
		redisClient: redis.NewClient(&redis.Options{
			Addr: "host.docker.internal:6379", // TODO: make configurable, got so many options to worry about now
		}),
		watchService: watch.New(ctx),
	}, nil
}

type APIServer struct {
	ctx          context.Context
	cancel       context.CancelFunc
	server       *http.Server
	redisClient  *redis.Client
	watchService *watch.WatchService
}
