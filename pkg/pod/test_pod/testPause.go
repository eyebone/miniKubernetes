package test_pod

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"regexp"
)

const (
	PauseImage = "registry.cn-hangzhou.aliyuncs.com/google_containers/pause-amd64:3.1"
)

func createPause(pod *Pod) (string, string, error) {
	fmt.Printf("========starting create pause container=========\n")

	fmt.Printf("============ begin to deal with exposed ports =============\n")
	ports := make(map[nat.Port]struct{})
	portBindings := make(map[nat.Port][]nat.PortBinding)

	for _, container := range pod.configs.Spec.Containers {
		for _, port := range container.Ports {
			if port.ContainerPort == 0 {
				return "", "", fmt.Errorf("container port 0 is not valid")
			}
			portStr := fmt.Sprintf("%d", port.ContainerPort)
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

	// 生成有效的 pause 容器名称
	pauseContainerName := fmt.Sprintf("%s-pause", pod.configs.Metadata.Name)
	//pauseContainerName := pod.configs.Metadata.Name + "pause"
	// 确保生成的容器名称符合 Docker 的命名规则
	re := regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_.-]*`)
	if !re.MatchString(pauseContainerName) || pauseContainerName == "" {
		return "", "", fmt.Errorf("invalid container name: %s", pauseContainerName)
	}

	configOptions := &container.Config{
		Image:        PauseImage,
		ExposedPorts: ports,
	}
	hostConfig := &container.HostConfig{
		PortBindings:    portBindings,
		PublishAllPorts: true,
		IpcMode:         "shareable",
		RestartPolicy: container.RestartPolicy{
			Name: "always", // 设置重启策略为"always"，容器将总是自动重启
			// 可选的重启策略：
			// - "no"：无重启策略
			// - "always"：容器总是自动重启
			// - "on-failure"：容器在非零退出状态时重启（默认最多重启3次）
			// - "unless-stopped"：除非手动停止，否则容器总是自动重启
		},
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
