package pod

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	PauseImage = "registry.cn-hangzhou.aliyuncs.com/google_containers/pause-amd64:3.1"
)

func CreatePauseContainer(pod *Pod) (string, string, error) {
	fmt.Printf("========starting create pause container=========\n")

	fmt.Printf("============ begin to deal with exposed ports =============\n")
	ports := make(map[nat.Port]struct{})
	portBindings := make(map[nat.Port][]nat.PortBinding)

	for _, container := range pod.Containers {
		for _, port := range container.PortMappings {
			if port.ContainerPort == 0 {
				return "", "", fmt.Errorf("container port 0 is not valid")
			}
			portStr := fmt.Sprintf("%d/tcp", port.ContainerPort)
			natPort, err := nat.NewPort("tcp", portStr)
			if err != nil {
				return "", "", fmt.Errorf("failed to parse port: %w", err)
			}
			ports[natPort] = struct{}{}
			portBindings[natPort] = []nat.PortBinding{
				{
					HostPort: fmt.Sprintf("%d", port.HostPort),
				},
			}
		}
	}

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	pauseContainerName := pod.Name + "-pause"
	configOptions := &container.Config{
		Image:        PauseImage,
		ExposedPorts: ports,
	}
	hostConfig := &container.HostConfig{
		PortBindings:    portBindings,
		PublishAllPorts: true,
		IpcMode:         "shareable",
	}

	response, err := cli.ContainerCreate(context.Background(), configOptions, hostConfig, nil, nil, pauseContainerName)
	if err != nil {
		return "", "", fmt.Errorf("failed to create pause container: %w", err)
	}

	err = cli.ContainerStart(context.Background(), response.ID, container.StartOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to start pause container: %w", err)
	}

	return response.ID, pauseContainerName, nil
}
