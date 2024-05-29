package pod

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/client"
)

type PodManager struct {
	pods map[string]*Pod
	mu   sync.Mutex
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

func (pm *PodManager) GetPod(name string) (*Pod, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pod, exists := pm.pods[name]
	return pod, exists
}

func (pm *PodManager) StartPod(ctx context.Context, cli *client.Client, p *Pod) {
	p.Start(ctx, cli)
	pm.AddPod(p)
}

func (pm *PodManager) StopPod(ctx context.Context, cli *client.Client, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pod, exists := pm.pods[name]
	if !exists {
		return fmt.Errorf("Pod not found: %s", name)
	}
	pod.Stop(ctx, cli)
	return nil
}

func (pm *PodManager) RemovePod(ctx context.Context, cli *client.Client, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pod, exists := pm.pods[name]
	if !exists {
		return fmt.Errorf("Pod not found: %s", name)
	}
	pod.Remove(ctx, cli)
	delete(pm.pods, name)
	return nil
}

func (pm *PodManager) DisplayPodStatus(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pod, exists := pm.pods[name]
	if !exists {
		return fmt.Errorf("Pod not found: %s", name)
	}
	pod.DisplayStatus()
	return nil
}

func (pm *PodManager) Monitor(ctx context.Context, cli *client.Client) {
	for {
		pm.mu.Lock()
		for _, pod := range pm.pods {
			go pod.Monitor(ctx, cli)
		}
		pm.mu.Unlock()
		time.Sleep(10 * time.Second) // Check every 10 seconds
	}
}
