package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"k8s/pkg/pod"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: [start | stop | remove | get | describe] <pod yaml file or pod name>")
		return
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	command := os.Args[1]
	arg := os.Args[2]

	podManager := pod.NewPodManager()

	// 启动监控所有Pod的协程
	go podManager.Monitor(ctx, cli)

	// 加载现有的Pod配置
	loadExistingPods(podManager, cli)

	switch command {
	case "start":
		absPath, err := filepath.Abs(arg)
		if err != nil {
			log.Fatalf("Error getting absolute path: %v", err)
		}
		p, err := pod.NewPod(absPath)
		if err != nil {
			log.Fatalf("Error creating pod: %v", err)
		}
		podManager.StartPod(ctx, cli, p)
	case "stop":
		if err := podManager.StopPod(ctx, cli, arg); err != nil {
			log.Fatalf("Error stopping pod: %v", err)
		}
	case "remove":
		if err := podManager.RemovePod(ctx, cli, arg); err != nil {
			log.Fatalf("Error removing pod: %v", err)
		}
	case "get":
		if err := podManager.DisplayPodStatus(arg); err != nil {
			log.Fatalf("Error getting pod status: %v", err)
		}
	case "describe":
		if err := podManager.DisplayPodStatus(arg); err != nil {
			log.Fatalf("Error describing pod: %v", err)
		}
	default:
		fmt.Println("Unknown command. Usage: [start | stop | remove | get | describe] <pod yaml file or pod name>")
	}
}

// 从yaml文件加载现有的Pod
func loadExistingPods(pm *pod.PodManager, cli *client.Client) {
	files, err := filepath.Glob("*.yaml")
	if err != nil {
		log.Fatalf("Error loading pod configurations: %v", err)
	}

	for _, file := range files {
		p, err := pod.NewPod(file)
		if err != nil {
			log.Printf("Error creating pod from file %s: %v", file, err)
			continue
		}
		pm.AddPod(p)
	}
}
