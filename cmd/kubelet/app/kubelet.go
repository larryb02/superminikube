package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"superminikube/kubelet"
	"superminikube/kubelet/runtime"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubelet",
		Short: "Node agent, sole purpose is running and maintaining pods",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	}

	return cmd
}

func Run() {
	slog.Info("Starting Kubelet...")
	slog.Debug("just a test")
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()
	// TODO: create runtime inside agent
	rt, err := runtime.NewDockerRuntime()
	if err != nil {
		slog.Error("Failed to start kubelet: ", "error", err)
		os.Exit(1)
	}
	kubelet, err := kubelet.NewKubelet(rt, ctx)
	if err != nil {
		slog.Error("Failed to start Kubelet:", "error", err)
		os.Exit(1)
	}
	fmt.Printf("checking something: %v", err)
	<-ctx.Done()
	kubelet.Cleanup()
}
