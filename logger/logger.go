package logger

import (
	"log/slog"
	"os"
)

var h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
var Logger = slog.New(h)
