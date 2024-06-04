package test_pod

// test pod create
import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"new_k8s/etcd"
	"new_k8s/pkg/container"
	"strconv"
	"strings"
	"time"
)

type PodPhase string

const (
	Pending   PodPhase = "Pending"
	Running   PodPhase = "Running"
	Succeeded PodPhase = "Succeeded"
	Failed    PodPhase = "Failed"
	Unknown   PodPhase = "Unknown"
)

// 定义结构体以匹配 YAML 文件的结构
type Pod struct {
	Configs struct {
		Kind     string   `yaml:"kind" json:"kind"`
		Metadata Metadata `yaml:"metadata" json:"metadata"`
		Spec     Spec     `yaml:"spec" json:"spec"`
	}
	StartTime      time.Time            `json:"StartTime"`
	PodPhase       PodPhase             `json:"PodPhase"`
	PauseContainer *container.Container `json:"PauseContainer"`
}

type Metadata struct {
	Name string `yaml:"name" json:"name"`
}

type Spec struct {
	Containers []container.Container `yaml:"containers" json:"containers"`
	Volumes    []Volume              `yaml:"volumes" json:"volumes"`
}

type Volume struct {
	Name     string   `yaml:"name" json:"name"`
	HostPath HostPath `yaml:"hostPath" json:"hostPath"`
}

type HostPath struct {
	Path string `yaml:"path" json:"path"`
}

func NewPod(yamlFilePath string, etcdclient etcd.MyEtcdClient) (Pod, string, error) {
	// 读取 YAML 文件
	data, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return Pod{PodPhase: Unknown}, "error new pod", fmt.Errorf("Error reading YAML file %s: %v", yamlFilePath, err)
	}

	// 解析 YAML 文件
	var pod Pod
	pod = Pod{
		PodPhase: Pending,
	}
	err = yaml.Unmarshal(data, &pod.Configs)
	if err != nil {
		return Pod{PodPhase: Unknown}, "error parsing yaml", fmt.Errorf("Error parsing YAML file %s: %v", yamlFilePath, err)
	}

	for _, c := range pod.Configs.Spec.Containers {
		var cont container.Container
		cont.Name = pod.Configs.Metadata.Name + "-" + c.Name // 容器 名字为pod名+容器本身名
		cont.Image = c.Image
		cont.Command = c.Command
		cont.Args = c.Args
		cont.Resources.Limits.CPU = c.Resources.Limits.CPU
		// 处理资源限制
		cont.Resources.Limits.Memory = c.Resources.Limits.Memory

		// 处理端口映射
		for _, p := range c.Ports {
			port := container.Port{
				ContainerPort: p.ContainerPort,
				HostPort:      p.HostPort,
			}
			cont.Ports = append(cont.Ports, port)

		}
		// 处理卷挂载

		vMount := container.VolumeMount{
			Name:      c.VolumeMount.Name,
			MountPath: c.VolumeMount.MountPath,
		}
		cont.VolumeMount = vMount

		pod.Configs.Spec.Containers = append(pod.Configs.Spec.Containers)
	}
	// pause container
	pod.PauseContainer = container.NewContainer()
	pod.PauseContainer.Image = PauseImage

	return pod, pod.Configs.Metadata.Name, nil
}

func parseCPU(cpu string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSuffix(cpu, "m"))
	return value, err
}

func parseMemory(memory string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSuffix(memory, "Mi"))
	return value, err
}
