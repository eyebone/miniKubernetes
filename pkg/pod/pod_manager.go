package pod

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/client"
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
	pm.pods[p.Name] = p
}

func (pm *PodManager) StartPod(ctx context.Context, cli *client.Client, p *Pod) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p.Start(ctx, cli)
	if p.PodStatus != Running {
		fmt.Printf("Failed to start pod: %v\n", p.PodStatus)
		return
	}
	fmt.Printf("Pod %s started successfully\n", p.Name)
	pm.pods[p.Name] = p
}

func (pm *PodManager) StopPod(ctx context.Context, cli *client.Client, podName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.pods[podName]
	if !exists {
		return fmt.Errorf("pod %s not found", podName)
	}
	p.Stop(ctx, cli)
	fmt.Printf("Pod %s stopped\n", podName)
	return nil
}

func (pm *PodManager) RemovePod(ctx context.Context, cli *client.Client, podName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.pods[podName]
	if !exists {
		return fmt.Errorf("pod %s not found", podName)
	}
	p.Remove(ctx, cli)
	fmt.Printf("Pod %s removed\n", podName)
	delete(pm.pods, podName)
	return nil
}

func (pm *PodManager) DisplayPodStatus(podName string) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.pods[podName]
	if !exists {
		return fmt.Errorf("pod %s not found", podName)
	}
	p.DisplayStatus()
	return nil
}

func (pm *PodManager) Monitor(ctx context.Context, cli *client.Client) {
	for {
		pm.mu.RLock()
		for _, p := range pm.pods {
			p.Monitor(ctx, cli)
		}
		pm.mu.RUnlock()
	}
}
