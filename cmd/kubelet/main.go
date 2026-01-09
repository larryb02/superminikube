package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"superminikube/pkg/kubelet"

	"github.com/spf13/cobra"
)

// TODO: Return error in Run
func NewAgentCommand() *cobra.Command {
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
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()
	// TODO: more things that need to be configured
	// how do i plan on generating nodenames
	// maybe i should verify nodes exist as well on server side
	k, err := kubelet.NewKubelet("http://localhost:8080", "agent-node-0")
	if err != nil {
		slog.Error("Failed to start Kubelet:", "error", err)
		os.Exit(1)
	}
	err = k.Start(ctx)
	if err != nil {
		slog.Error("Failed to start Kubelet:", "error", err)
		os.Exit(1)
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	cmd := NewAgentCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
