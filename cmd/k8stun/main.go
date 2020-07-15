package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"k8s.io/client-go/tools/clientcmd"

	"k8stun"
)

var (
	kubeConfig = "~/.kube/config"
	configFile string
)

func waitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func main() {

	if os.Getenv("KUBECONFIG") != "" {
		kubeConfig = os.Getenv("KUBECONFIG")
	}

	flag.StringVar(
		&kubeConfig,
		"k8sconfig",
		kubeConfig,
		"Path to Kubernetes config file")
	flag.StringVar(
		&configFile,
		"config",
		configFile,
		"Path to k8stun config file")
	flag.Parse()

	if configFile == "" || kubeConfig == "" {
		flag.PrintDefaults()
		return
	}

	cfg, err := k8stun.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("error loading config: %s\n", err)
		panic(err)
	}

	// uses the current context in kubeconfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Panicf("error: %s", err.Error())
	}

	// Start service
	svc := k8stun.NewService(restConfig, cfg.Tunnels)
	svc.Start()
	defer svc.Stop()

	// Run until canceled
	waitForInterrupt()
}
