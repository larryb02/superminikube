package main

import (
	"log"
	"os"
	"superminikube/kubelet"
)

func main() {
	runtime, err := kubelet.NewDockerRuntime()
	if err != nil {
		log.Fatalf("Failed to establish connection to Docker")
	}
	err = runtime.Ping()
	if err != nil {
		log.Fatalf("Failed to ping %v", err)
	}
	log.Println("Successfully pinged Docker")
	args := os.Args[1:]
	err = runtime.Pull(args[0])
	if err != nil {
		log.Printf("Failed to pull %s: %v", args[0], err)
	}
	id, err := runtime.CreateContainer()
	if err != nil {
		log.Printf("%v", err)
	}
	err = runtime.StartContainer(id)
	if err != nil {
		log.Printf("%v", err)
	}
	err = runtime.StopContainer(id)
	if err != nil {
		log.Printf("%v", err)
	}
	err = runtime.RemoveContainer(id)
	if err != nil {
		log.Printf("%v", err)
	}
}
