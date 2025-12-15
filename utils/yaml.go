package main

import (
	"log/slog"
	"os"

	"github.com/moby/moby/api/types/container"
	"gopkg.in/yaml.v3"
)


func main() {
	var cfg container.Config
	data := []byte{}
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		slog.Error("Failed to Marshal: ", "error", err)
		os.Exit(1)
	}
}
