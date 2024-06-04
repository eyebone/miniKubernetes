package test_pod

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"new_k8s/etcd" // 导入etcd包
	"regexp"
)

const (
	PauseImage = "registry.cn-hangzhou.aliyuncs.com/google_containers/pause-amd64:3.1"
)

// PauseContainerMeta 定义Pause容器的元数据信息结构
type PauseContainerMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Pod  string `json:"pod"`
}

func createPause(pod *Pod, etcdClient etcd.MyEtcdClient) (string, string, error) {
	fmt.Printf("========starting create pause container=========\n")

	fmt.Printf("============ begin to deal with exposed ports =============\n")
	ports := make(map[nat.Port]struct{})
	portBindings := make(map[nat.Port][]nat.PortBinding)

	for _, container := range pod.Configs.Spec.Containers {
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
	// 这里的cli是dockerclient
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	// 生成有效的 pause 容器名称
	pauseContainerName := fmt.Sprintf("%s-pause", pod.Configs.Metadata.Name)
	// 确保生成的容器名称符合 Docker 的命名规则
	re := regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_.-]*`)
	if !re.MatchString(pauseContainerName) || pauseContainerName == "" {
		return "", "", fmt.Errorf("invalid container name: %s", pauseContainerName)
	}
	fmt.Println("pause container name: ", pauseContainerName, "\n")

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

	// 创建 Pause 容器的元数据结构体
	pauseMeta := PauseContainerMeta{
		ID:   response.ID,
		Name: pauseContainerName,
		Pod:  pod.Configs.Metadata.Name,
	}

	// 序列化 Pause 容器元数据为 JSON 字符串
	pauseMetaData, err := json.Marshal(pauseMeta)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal pause container metadata: %w", err)
	}

	// 将 Pause 容器元数据写入 etcd
	key := fmt.Sprintf("pods/%s/pause", pod.Configs.Metadata.Name)
	if err := etcdClient.Put(key, string(pauseMetaData)); err != nil {
		return "", "", fmt.Errorf("failed to write pause container metadata to etcd: %w", err)
	}

	return response.ID, pauseContainerName, nil
}
