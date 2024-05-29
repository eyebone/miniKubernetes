package pod

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
	"k8s/pkg/container"
)

type PodStatus string

const (
	Pending   PodStatus = "Pending"
	Running   PodStatus = "Running"
	Succeeded PodStatus = "Succeeded"
	Failed    PodStatus = "Failed"
	Unknown   PodStatus = "Unknown"
	Stopped   PodStatus = "Stopped"
)

type Volume struct {
	Name string
	Path string
}

type Pod struct {
	Kind           string
	Name           string
	Containers     []container.Container
	Volumes        []Volume
	StartTime      time.Time
	PodStatus      PodStatus
	PauseImage     string
	PauseContainer *container.Container
}

func NewPod(yamlFilePath string) (*Pod, error) {
	data, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	pod := &Pod{
		Kind:       config["kind"].(string),
		Name:       config["name"].(string),
		PodStatus:  Pending,
		PauseImage: "k8s.gcr.io/pause:3.5",
	}

	containersNode := config["containers"].([]interface{})
	for _, c := range containersNode {
		cMap := c.(map[interface{}]interface{})
		cont := container.NewContainer()
		cont.Name = fmt.Sprintf("%s-%s", pod.Name, cMap["name"].(string))
		cont.Image = cMap["image"].(string)

		if cmds, ok := cMap["command"]; ok {
			for _, cmd := range cmds.([]interface{}) {
				cont.Command = append(cont.Command, cmd.(string))
			}
		}

		if args, ok := cMap["args"]; ok {
			for _, arg := range args.([]interface{}) {
				cont.CmdArgs = append(cont.CmdArgs, arg.(string))
			}
		}

		if resources, ok := cMap["resources"]; ok {
			limits := resources.(map[interface{}]interface{})["limits"].(map[interface{}]interface{})
			if cpu, ok := limits["cpu"]; ok {
				cont.CPU = cpu.(int)
			}
			if memory, ok := limits["memory"]; ok {
				cont.Memory = memory.(int)
			}
		}

		if ports, ok := cMap["ports"]; ok {
			for _, p := range ports.([]interface{}) {
				port := p.(map[interface{}]interface{})
				portMapping := container.PortMapping{
					ContainerPort: port["containerPort"].(int),
					HostPort:      port["hostPort"].(int),
				}
				cont.PortMappings = append(cont.PortMappings, portMapping)
			}
		}

		if volumeMounts, ok := cMap["volumeMounts"]; ok {
			for _, v := range volumeMounts.([]interface{}) {
				volumeMount := v.(map[interface{}]interface{})
				cont.VolumeMounts[volumeMount["name"].(string)] = volumeMount["mountPath"].(string)
			}
		}

		pod.Containers = append(pod.Containers, *cont)
	}

	if volumes, ok := config["volumes"]; ok {
		for _, v := range volumes.([]interface{}) {
			volume := v.(map[interface{}]interface{})
			vol := Volume{
				Name: volume["name"].(string),
				Path: volume["hostPath"].(map[interface{}]interface{})["path"].(string),
			}
			pod.Volumes = append(pod.Volumes, vol)
		}
	}

	pod.PauseContainer = container.NewContainer()
	pod.PauseContainer.Name = fmt.Sprintf("%s-pause", pod.Name)
	pod.PauseContainer.Image = pod.PauseImage

	return pod, nil
}

func (p *Pod) Start(ctx context.Context, cli *client.Client) {
	p.StartTime = time.Now()
	p.PodStatus = Running
	// Start the pause container
	if err := p.PauseContainer.PullImage(ctx, cli); err != nil {
		fmt.Printf("Failed to pull pause container image: %v\n", err)
		p.PodStatus = Failed
		return
	}
	if err := p.PauseContainer.Start(ctx, cli); err != nil {
		fmt.Printf("Failed to start pause container: %v\n", err)
		p.PodStatus = Failed
		return
	}

	// Ensure volumes are properly mounted for inter-container communication
	allContainersRunning := true
	for i := range p.Containers {
		for _, volume := range p.Volumes {
			for name, mountPath := range p.Containers[i].VolumeMounts {
				if name == volume.Name {
					p.Containers[i].VolumeMounts[volume.Path] = mountPath
				}
			}
		}
		if err := p.Containers[i].PullImage(ctx, cli); err != nil {
			fmt.Printf("Failed to pull container image: %v\n", err)
			allContainersRunning = false
			break
		}
		if err := p.Containers[i].Start(ctx, cli); err != nil {
			fmt.Printf("Failed to start container: %v\n", err)
			allContainersRunning = false
			break
		}
	}

	if allContainersRunning {
		p.PodStatus = Running
	} else {
		p.PodStatus = Failed
	}
}

func (p *Pod) Stop(ctx context.Context, cli *client.Client) {
	allContainersStopped := true
	for i := range p.Containers {
		if err := p.Containers[i].Stop(ctx, cli); err != nil {
			fmt.Printf("Failed to stop container: %v\n", err)
			allContainersStopped = false
		} else {
			p.Containers[i].Status = "stopped"
		}
	}
	if err := p.PauseContainer.Stop(ctx, cli); err != nil {
		fmt.Printf("Failed to stop pause container: %v\n", err)
		allContainersStopped = false
	} else {
		p.PauseContainer.Status = "stopped"
	}

	if allContainersStopped {
		p.PodStatus = Stopped
	} else {
		p.PodStatus = Failed
	}
}

func (p *Pod) Remove(ctx context.Context, cli *client.Client) {
	allContainersRemoved := true
	for i := range p.Containers {
		if err := p.Containers[i].Remove(ctx, cli); err != nil {
			fmt.Printf("Failed to remove container: %v\n", err)
			allContainersRemoved = false
		} else {
			p.Containers[i].Status = "removed"
		}
	}
	if err := p.PauseContainer.Remove(ctx, cli); err != nil {
		fmt.Printf("Failed to remove pause container: %v\n", err)
		allContainersRemoved = false
	} else {
		p.PauseContainer.Status = "removed"
	}

	if allContainersRemoved {
		p.PodStatus = Succeeded
	} else {
		p.PodStatus = Failed
	}
}

func (p *Pod) DisplayStatus() {
	for _, c := range p.Containers {

		fmt.Printf("Container %s is currently %s.\n", c.Name, c.Status)
	}
	fmt.Printf("Pod Name: %s\n", p.Name)
	fmt.Printf("Pod Status: %s\n", p.PodStatus)
	p.DisplayRunTime()
}

func (p *Pod) DisplayRunTime() {
	duration := time.Since(p.StartTime)
	if p.PodStatus == Pending {
		//fmt.Printf("Pod %s is pending.\n", p.Name)
		fmt.Printf("Pod %s has been pendding for %s.\n", p.Name, duration.Truncate(time.Second))

	} else {
		fmt.Printf("Pod %s has been running for %s.\n", p.Name, duration.Truncate(time.Second))
	}
}

func (p *Pod) Monitor(ctx context.Context, cli *client.Client) {
	for {
		for i := range p.Containers {
			running, err := p.Containers[i].IsRunning(ctx, cli)
			if err != nil {
				fmt.Printf("Error checking container status: %v\n", err)
				p.PodStatus = Unknown
				continue
			}
			if !running {
				fmt.Printf("Container %s has crashed. Restarting...\n", p.Containers[i].Name)
				if err := p.Containers[i].Start(ctx, cli); err != nil {
					fmt.Printf("Failed to restart container: %v\n", err)
					p.PodStatus = Failed
				} else {
					p.PodStatus = Running
				}
			}
		}
		time.Sleep(10 * time.Second) // Check every 10 seconds
	}
}
