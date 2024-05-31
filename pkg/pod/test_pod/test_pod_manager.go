package test_pod

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"sync"
)

type PodManager struct {
	mu   sync.RWMutex
	pods map[string]*Pod
}

func NewPodManager() *PodManager {
	return &PodManager{
		pods: make(map[string]*Pod),
	}
}

func (pm *PodManager) AddPod(p *Pod) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pods[p.configs.Metadata.Name] = p
}

func (pm *PodManager) StartPod(ctx context.Context, cli *client.Client, p *Pod) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	p.Start(ctx, cli)
	if p.PodPhase != Running {
		fmt.Printf("Failed to start pod: %v\n", p.PodPhase)
		return
	}
	fmt.Printf("Pod %s started successfully\n", p.configs.Metadata.Name)
	pm.pods[p.configs.Metadata.Name] = p
}

//func main() {
//
//}
