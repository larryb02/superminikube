package main

import (
	"os"
	"superminikube/kubelet"
)

func main() {
	kubelet.Run(os.Args) // Note: Passing args is temporary, kubelet will eventually receive commands over http
}
