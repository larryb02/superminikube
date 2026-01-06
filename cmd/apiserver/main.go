package main

import (
	"fmt"
	"os"
	"log/slog"

	"github.com/spf13/cobra"
	"superminikube/pkg/apiserver"
)

func Run() {
	err := apiserver.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to start apiserver: %v", err))
		os.Exit(1)
	}
}

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiserver",
		Short: "apiserver",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	}

	return cmd
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	cmd := NewAPIServerCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
