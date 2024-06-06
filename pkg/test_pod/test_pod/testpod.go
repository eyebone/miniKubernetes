package test_pod

// test pod create
import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"new_k8s/pkg/container"
	"new_k8s/pkg/test_pod/random"
	"new_k8s/tools/etcd/etcd"
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
	StartTime      time.Time            `json:"StartTime" yaml:"StartTime"`
	PodPhase       PodPhase             `json:"PodPhase" yaml:"PodPhase"`
	PauseContainer *container.Container `json:"PauseContainer" yaml:"PauseContainer"`
	PodIP          string               `json:"PodIP" yaml:"PodIP"`
}

type Metadata struct {
	Name string `yaml:"name" json:"name"`
	Uid  string `yaml:"uid,omitempty" json:"uid,omitempty"`
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

// 创建和初试化一个新的pod
// 状态为pending
func NewPod(yamlFilePath string, etcdclient etcd.MyEtcdClient) (Pod, error) {
	// 读取 YAML 文件
	data, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return Pod{PodPhase: Unknown}, fmt.Errorf("Error reading YAML file %s: %v", yamlFilePath, err)
	}

	// 解析 YAML 文件
	var pod Pod
	pod = Pod{
		PodPhase: Pending,
	}
	err = yaml.Unmarshal(data, &pod.Configs)
	if err != nil {
		return Pod{PodPhase: Unknown}, fmt.Errorf("Error parsing YAML file %s: %v", yamlFilePath, err)
	}

	// 设定pod的唯一UID
	pod.Configs.Metadata.Uid = random.GenerateXid()

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
	// pod ip
	pod.PodIP = ""
	return pod, nil
}
