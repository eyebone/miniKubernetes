package container

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	_ "log"
	"os"
	"regexp"
	"strconv"
	_ "strconv"
	"strings"
	_ "strings"
)

type Container struct {
	Name        string      `yaml:"name"`
	Image       string      `yaml:"image"`
	Ports       []Port      `yaml:"ports"`
	VolumeMount VolumeMount `yaml:"volumeMounts"`
	Command     []string    `yaml:"command"`
	Args        []string    `yaml:"args""`
	Resources   Resources   `yaml:"resources"`
	Status      string
}

type Resources struct {
	Limits struct {
		Memory string `yaml:"memory"`
		CPU    string `yaml:"cpu"`
	}
}

type Port struct {
	ContainerPort int `yaml:"containerPort"`
	HostPort      int `yaml:"hostPort"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
}

func NewContainer() *Container {
	return &Container{}
}

func EnsureImage(targetImage string) error {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	defer cli.Close()

	exist, err := ImageExist(targetImage)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	fmt.Printf("image %s doesn't exist, automatically pulling\n", targetImage)
	reader, err := cli.ImagePull(context.Background(), targetImage, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", targetImage, err)
	}
	io.Copy(os.Stdout, reader)

	return nil
}

func ImageExist(targetImage string) (bool, error) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == targetImage {
				fmt.Printf("======> have found the image <======\n")
				return true, nil
			}
		}
	}

	return false, nil
}

// 创建容器
func CreateContainer(c *Container, pauseMode string, volumeBinds []string) (string, string, error) {
	ctx := context.Background()

	cpu, err := parseCPU(c.Resources.Limits.CPU)
	if err != nil {
		fmt.Printf("Failed to parse CPU for container %s: %v\n", c.Name, err)
	}

	memory, err := parseMemory(c.Resources.Limits.Memory)
	if err != nil {
		fmt.Printf("Failed to parse memory for container %s: %v\n", c.Name, err)
	}

	fmt.Printf("volume binds: %v\n", volumeBinds)

	// 定义正则表达式来匹配路径
	re := regexp.MustCompile(`volume:(\/[^\]]+)`)

	//hostPath := ""
	// 遍历数组并提取路径
	for _, bind := range volumeBinds {
		// 查找匹配的子字符串
		matches := re.FindStringSubmatch(bind)

		if len(matches) > 1 {
			// 提取路径
			path := matches[1]
			fmt.Printf("Extracted path: %s\n", path)
			//hostPath = path
		} else {
			fmt.Println("No path found")
		}
	}

	hostConfig := &container.HostConfig{
		PidMode:     container.PidMode(pauseMode),
		IpcMode:     container.IpcMode(pauseMode),
		NetworkMode: container.NetworkMode(pauseMode),
		Binds:       volumeBinds,
		Resources: container.Resources{
			CPUPeriod: 1000000,
			CPUQuota:  int64(cpu * 1000), // Convert milli cores to microseconds
			Memory:    int64(memory) * 1024 * 1024,
		},
	}
	fmt.Printf("volumeMount in container: %v\n", c.VolumeMount)

	// Do not set port bindings and exposed ports if the container is sharing network with pause container
	portBindings := make(map[nat.Port][]nat.PortBinding)

	exposedPorts := make(map[nat.Port]struct{})
	if pauseMode == "" {
		for _, pm := range c.Ports {
			port := nat.Port(fmt.Sprintf("%d/tcp", pm.ContainerPort))
			exposedPorts[port] = struct{}{}
			portBindings[port] = []nat.PortBinding{
				{
					HostPort: fmt.Sprintf("%d", pm.HostPort),
				},
			}
		}
		hostConfig.PortBindings = portBindings
	}
	cli, _ := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        c.Image,
		Tty:          true,
		Cmd:          append(c.Command, c.Args...),
		ExposedPorts: exposedPorts,
	}, hostConfig, nil, nil, c.Name)

	if err != nil {
		return "", "", fmt.Errorf("failed to create container: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	fmt.Println("create container response ", string(respBytes))

	return resp.ID, c.Name, nil
}

func StartContainer(containerID string) error {

	cli, _ := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	defer cli.Close()

	err := cli.ContainerStart(context.Background(), containerID, container.StartOptions{})

	if err != nil {
		return err
	}
	fmt.Println("container: ", containerID, "start successfully!\n")
	//// wait running
	//statusCh, errCh := cli.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)
	//select {
	//case err := <-errCh:
	//	fmt.Printf("container wait error: %v\n", err)
	//	return err
	//
	//	break
	//case status := <-statusCh:
	//	fmt.Println("container start success ", status.StatusCode)
	//}
	//
	//// 获取日志
	//out, err := cli.ContainerLogs(context.Background(), containerID, container.LogsOptions{ShowStdout: true})
	//if err != nil {
	//	fmt.Printf("get container log error: %v\n", err)
	//	return err
	//}
	//io.Copy(os.Stdout, out)
	return nil
}

func (c *Container) Stop(ctx context.Context, cli *client.Client) error {
	timeout := 10 // Timeout in seconds

	if err := cli.ContainerStop(ctx, c.Name, container.StopOptions{Timeout: &timeout}); err != nil {
		return err
	} else {
		fmt.Printf("container %s is stopped.\n", c.Name)
		c.Status = "stopped"
		return nil
	}
}

func (c *Container) Remove(ctx context.Context, cli *client.Client) error {
	if err := cli.ContainerRemove(ctx, c.Name, container.RemoveOptions{Force: true}); err != nil {
		return err
	} else {
		fmt.Printf("container %s is removed.\n", c.Name)
		c.Status = "removed"
		return nil
	}
}

func (c *Container) IsRunning(ctx context.Context, cli *client.Client) (bool, error) {
	json, err := cli.ContainerInspect(ctx, c.Name)
	if err != nil {
		return false, err
	}
	return json.State.Running, nil
}

func parseCPU(cpu string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSuffix(cpu, "m"))
	return value, err
}

func parseMemory(memory string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSuffix(memory, "Mi"))
	return value, err
}
