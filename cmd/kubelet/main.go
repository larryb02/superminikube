package main

import (
	"log/slog"
	"os"

	"superminikube/kubelet"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	kubelet.Run(os.Args) // Note: Passing args is temporary, kubelet will eventually receive commands over http
}
