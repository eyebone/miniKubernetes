package test_pod

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	//"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"net"
	"new_k8s/pkg/test_pod/random"
	"new_k8s/tools/etcd/etcd"
	"new_k8s/tools/flannel"
	"regexp"
)

// NOTE: 这是pause容器镜像
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
	//fmt.Printf("========starting create pause container=========\n")

	//fmt.Printf("============ begin to deal with exposed ports =============\n")
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
	pauseID := random.GenerateXid()
	pauseContainerName := fmt.Sprintf("%s-pauseContainer-%s", pod.Configs.Metadata.Name, pauseID)
	// 确保生成的容器名称符合 Docker 的命名规则
	re := regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_.-]*`)
	if !re.MatchString(pauseContainerName) || pauseContainerName == "" {
		return "", "", fmt.Errorf("invalid container name: %s", pauseContainerName)
	}
	// TODO: 分配子网中的一个IP地址
	//cxt := context.Background()
	//ip, err := AllocateIP(&etcdClient, cxt)
	if err != nil {
		return "", "", fmt.Errorf("failed to allocate IP: %w", err)
	}
	// TODO: 自定义网络

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
	// TODO: 添加pause 容器的networkingConfig
	//networkingConfig := &network.NetworkingConfig{
	//	EndpointsConfig: map[string]*network.EndpointSettings{
	//		"bridge": {
	//			IPAMConfig: &network.EndpointIPAMConfig{
	//				IPv4Address: ip,
	//			},
	//		},
	//	},
	//}

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
	key := fmt.Sprintf("pod-pausecontainers/%s", pod.Configs.Metadata.Name)
	if err := etcdClient.Put(key, string(pauseMetaData)); err != nil {
		return "", "", fmt.Errorf("failed to write pause container metadata to etcd: %w", err)
	}

	return response.ID, pauseContainerName, nil
}

// TODO:
func AllocateIP(cli *etcd.MyEtcdClient, ctx context.Context) (string, error) {
	config, err := GetFlannelConfig(*cli, ctx)
	if err != nil {
		return "get flannel config failed\n", err
	}
	subnetMinIP := net.ParseIP(config.SubnetMin)
	subnetMaxIP := net.ParseIP(config.SubnetMax)
	for ip := subnetMinIP; !ip.Equal(subnetMaxIP); inc(ip) {
		ipStr := ip.String()
		allocated, err := IsIPAllocated(ctx, cli, ipStr)
		if err != nil {
			return "", err
		}
		if !allocated {
			err = allocate(ctx, ipStr, cli)
			if err != nil {
				return "", err
			}
			return ipStr, nil
		}
	}
	return "", fmt.Errorf("no available IP in subnet range %s - %s", config.SubnetMin, config.SubnetMax)

}

func GetFlannelConfig(cli etcd.MyEtcdClient, ctx context.Context) (flannel.FlannelConfig, error) {
	// 从etcd获得子网信息，存储在 /coreos.com/network/config
	config := flannel.FlannelConfig{}
	resp, err := cli.Client.Get(ctx, "/coreos.com/network/config")
	if err != nil {
		return config, err
	}
	config, err = flannel.MyFlannelMarshal(resp.Kvs[0].Value)
	return config, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func IsIPAllocated(ctx context.Context, cli *etcd.MyEtcdClient, ip string) (bool, error) {
	resp, err := cli.Client.Get(ctx, fmt.Sprintf("/coreos.com/network/allocated/%s", ip))
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) > 0, nil
}

func allocate(ctx context.Context, ip string, cli *etcd.MyEtcdClient) error {
	_, err := cli.Client.Put(ctx, fmt.Sprintf("/coreos.com/network/allocated/%s", ip), "allocated")
	return err
}
