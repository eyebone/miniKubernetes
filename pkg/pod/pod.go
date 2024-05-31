package pod

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"new_k8s/pkg/container"
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
	configs struct {
		Kind     string   `yaml:"kind"`
		Metadata Metadata `yaml:"metadata"`
		Spec     Spec     `yaml:"spec"`
	}
	StartTime time.Time
	PodPhase  PodPhase
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Spec struct {
	Containers []Container `yaml:"containers"`
	Volumes    []Volume    `yaml:"volumes"`
}

type Container struct {
	container.Container
}

type Volume struct {
	Name     string   `yaml:"name"`
	HostPath HostPath `yaml:"hostPath"`
}

type HostPath struct {
	Path string `yaml:"path"`
}

func NewPod(yamlFilePath string) (Pod, error) {
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
	err = yaml.Unmarshal(data, &pod.configs)
	if err != nil {
		return Pod{PodPhase: Unknown}, fmt.Errorf("Error parsing YAML file %s: %v", yamlFilePath, err)
	}
	// c 是解析yaml文件中的每一个container
	for _, c := range pod.configs.Spec.Containers {
		// cont是自定义类型的container
		var cont container.Container
		cont.Name = c.Name
		cont.Image = c.Image
		cont.Command = c.Command
		cont.Args = c.Args
		// 处理资源限制
		cont.Resources.Limits.CPU = c.Resources.Limits.CPU
		cont.Resources.Limits.Memory = c.Resources.Limits.Memory
		// 处理端口映射
		for _, p := range c.Ports {
			cont.Ports = append(cont.Ports, p)
		}
		// 处理卷挂载
		for _, vm := range c.VolumeMounts {
			cont.VolumeMounts = append(cont.VolumeMounts, vm)
		}
		pod.configs.Spec.Containers = append(pod.configs.Spec.Containers, Container{cont})
	}

	return pod, nil
}

//package pod
//
//import (
//	_ "context"
//	"fmt"
//	"io/ioutil"
//	"strconv"
//	"strings"
//	"time"
//
//	"gopkg.in/yaml.v2"
//	"new_k8s/pkg/container"
//)
//
//type PodStatus string
//
//const (
//	Pending   PodStatus = "Pending"
//	Running   PodStatus = "Running"
//	Succeeded PodStatus = "Succeeded"
//	Failed    PodStatus = "Failed"
//	Unknown   PodStatus = "Unknown"
//)
//
//type Volume struct {
//	Name string
//	Path string
//}
//
//type Pod struct {
//	Kind           string
//	Name           string
//	Containers     []container.Container
//	Volumes        []Volume
//	StartTime      time.Time
//	PodStatus      PodStatus
//	PauseImage     string
//	PauseContainer *container.Container
//}
//
//func NewPod(yamlFilePath string) (*Pod, error) {
//	data, err := ioutil.ReadFile(yamlFilePath)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read YAML file: %w", err)
//	}
//
//	var config struct {
//		Kind string `yaml:"kind"`
//		Name string `yaml:"name"`
//		Spec struct {
//			Containers []struct {
//				Name         string                  `yaml:"name"`
//				Image        string                  `yaml:"image"`
//				Command      []string                `yaml:"command"`
//				Args         []string                `yaml:"args"`
//				Ports        []container.PortMapping `yaml:"ports"`
//				VolumeMounts []struct {
//					Name      string `yaml:"name"`
//					MountPath string `yaml:"mountPath"`
//				} `yaml:"volumeMounts"`
//				Resources struct {
//					Limits struct {
//						Memory string `yaml:"memory"`
//						CPU    string `yaml:"cpu"`
//					} `yaml:"limits"`
//				} `yaml:"resources"`
//			} `yaml:"containers"`
//			Volumes []struct {
//				Name     string `yaml:"name"`
//				HostPath struct {
//					Path string `yaml:"path"`
//				} `yaml:"hostPath"`
//			} `yaml:"volumes"`
//		} `yaml:"spec"`
//		Status struct {
//			Phase string `yaml:"phase"`
//		} `yaml:"status"`
//	}
//
//	if err := yaml.Unmarshal(data, &config); err != nil {
//		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
//	}
//
//	pod := &Pod{
//		Kind:       config.Kind,
//		Name:       config.Name,
//		PodStatus:  Pending,
//		PauseImage: PauseImage,
//	}
//
//	fmt.Printf("Parsed YAML:\nKind: %s\nName: %s\n", pod.Kind, pod.Name)
//
//	for _, c := range config.Spec.Containers {
//		cont := container.NewContainer()
//		cont.Name = fmt.Sprintf("%s-%s", pod.Name, c.Name)
//		cont.Image = c.Image
//		cont.Command = c.Command
//		cont.CmdArgs = c.Args
//
//		for _, p := range c.Ports {
//			cont.PortMappings = append(cont.PortMappings, p)
//		}
//
//		for _, vm := range c.VolumeMounts {
//			cont.VolumeMounts[vm.MountPath] = vm.Name
//		}
//
//		if cpu, err := parseCPU(c.Resources.Limits.CPU); err == nil {
//			cont.CPU = cpu
//		} else {
//			fmt.Printf("Failed to parse CPU for container %s: %v\n", c.Name, err)
//		}
//
//		if memory, err := parseMemory(c.Resources.Limits.Memory); err == nil {
//			cont.Memory = memory
//		} else {
//			fmt.Printf("Failed to parse memory for container %s: %v\n", c.Name, err)
//		}
//
//		pod.Containers = append(pod.Containers, *cont)
//	}
//
//	for _, v := range config.Spec.Volumes {
//		vol := Volume{
//			Name: v.Name,
//			Path: v.HostPath.Path,
//		}
//		pod.Volumes = append(pod.Volumes, vol)
//	}
//
//	pod.PauseContainer = container.NewContainer()
//	pod.PauseContainer.Name = fmt.Sprintf("%s-pause", pod.Name)
//	pod.PauseContainer.Image = pod.PauseImage
//
//	fmt.Printf("Final Pod Data:\n%+v\n", pod)
//
//	return pod, nil
//}
//func parseCPU(cpu string) (int, error) {
//	value, err := strconv.Atoi(strings.TrimSuffix(cpu, "m"))
//	return value, err
//}
//
//func parseMemory(memory string) (int, error) {
//	value, err := strconv.Atoi(strings.TrimSuffix(memory, "Mi"))
//	return value, err
//}

//package pod
//
//import (
//	_ "context"
//	"fmt"
//	"io/ioutil"
//	"strconv"
//	"strings"
//	"time"
//
//	"gopkg.in/yaml.v2"
//	"new_k8s/pkg/container"
//)
//
//type PodStatus string
//
//const (
//	Pending   PodStatus = "Pending"
//	Running   PodStatus = "Running"
//	Succeeded PodStatus = "Succeeded"
//	Failed    PodStatus = "Failed"
//	Unknown   PodStatus = "Unknown"
//)
//
//type Volume struct {
//	Name string
//	Path string
//}
//
//type Pod struct {
//	Kind           string
//	Name           string
//	Containers     []container.Container
//	Volumes        []Volume
//	StartTime      time.Time
//	PodStatus      PodStatus
//	PauseImage     string
//	PauseContainer *container.Container
//}
//
//func NewPod(yamlFilePath string) (*Pod, error) {
//	data, err := ioutil.ReadFile(yamlFilePath)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read YAML file: %w", err)
//	}
//
//	var config struct {
//		Kind string `yaml:"kind"`
//		Name string `yaml:"name"`
//		Spec struct {
//			Containers []struct {
//				Name         string                  `yaml:"name"`
//				Image        string                  `yaml:"image"`
//				Command      []string                `yaml:"command"`
//				Args         []string                `yaml:"args"`
//				Ports        []container.PortMapping `yaml:"ports"`
//				VolumeMounts []struct {
//					Name      string `yaml:"name"`
//					MountPath string `yaml:"mountPath"`
//				} `yaml:"volumeMounts"`
//				Resources struct {
//					Limits struct {
//						Memory string `yaml:"memory"`
//						CPU    string `yaml:"cpu"`
//					} `yaml:"limits"`
//				} `yaml:"resources"`
//			} `yaml:"containers"`
//			Volumes []struct {
//				Name     string `yaml:"name"`
//				HostPath struct {
//					Path string `yaml:"path"`
//				} `yaml:"hostPath"`
//			} `yaml:"volumes"`
//		} `yaml:"spec"`
//		Status struct {
//			Phase string `yaml:"phase"`
//		} `yaml:"status"`
//	}
//
//	if err := yaml.Unmarshal(data, &config); err != nil {
//		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
//	}
//
//	pod := &Pod{
//		Kind:       config.Kind,
//		Name:       config.Name,
//		PodStatus:  Pending,
//		PauseImage: PauseImage,
//	}
//
//	fmt.Printf("Parsed YAML:\nKind: %s\nName: %s\n", pod.Kind, pod.Name)
//
//	for _, c := range config.Spec.Containers {
//		cont := container.NewContainer()
//		cont.Name = fmt.Sprintf("%s-%s", pod.Name, c.Name)
//		cont.Image = c.Image
//		cont.Command = c.Command
//		cont.CmdArgs = c.Args
//
//		for _, p := range c.Ports {
//			cont.PortMappings = append(cont.PortMappings, p)
//		}
//
//		for _, vm := range c.VolumeMounts {
//			cont.VolumeMounts[vm.MountPath] = vm.Name
//		}
//
//		if cpu, err := parseCPU(c.Resources.Limits.CPU); err == nil {
//			cont.CPU = cpu
//		} else {
//			fmt.Printf("Failed to parse CPU for container %s: %v\n", c.Name, err)
//		}
//
//		if memory, err := parseMemory(c.Resources.Limits.Memory); err == nil {
//			cont.Memory = memory
//		} else {
//			fmt.Printf("Failed to parse memory for container %s: %v\n", c.Name, err)
//		}
//
//		pod.Containers = append(pod.Containers, *cont)
//	}
//
//	for _, v := range config.Spec.Volumes {
//		vol := Volume{
//			Name: v.Name,
//			Path: v.HostPath.Path,
//		}
//		pod.Volumes = append(pod.Volumes, vol)
//	}
//
//	pod.PauseContainer = container.NewContainer()
//	pod.PauseContainer.Name = fmt.Sprintf("%s-pause", pod.Name)
//	pod.PauseContainer.Image = pod.PauseImage
//
//	fmt.Printf("Final Pod Data:\n%+v\n", pod)
//
//	return pod, nil
//}
//func parseCPU(cpu string) (int, error) {
//	value, err := strconv.Atoi(strings.TrimSuffix(cpu, "m"))
//	return value, err
//}
//
//func parseMemory(memory string) (int, error) {
//	value, err := strconv.Atoi(strings.TrimSuffix(memory, "Mi"))
//	return value, err
//}
