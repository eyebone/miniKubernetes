package pod

import (
	"context"
	"fmt"
	//"io"
	"new_k8s/pkg/container"
	//"os"
	"time"

	"github.com/docker/docker/client"
	//"github.com/docker/docker/api/types"
)

func (p *Pod) Start(ctx context.Context, cli *client.Client) {
	p.StartTime = time.Now()

	if err := container.EnsureImage(p.PauseImage); err != nil {
		fmt.Printf("Failed to pull pause container image: %v\n", err)
		p.PodStatus = Failed
		return
	}

	pauseID, pauseName, err := CreatePauseContainer(p)
	if err != nil {
		fmt.Printf("Failed to start pause container: %v\n", err)
		p.PodStatus = Failed
		return
	} else {
		p.PauseContainer.Name = pauseName
	}
	pauseMode := "container:" + pauseID
	allContainersRunning := true

	for i := range p.Containers {
		for _, volume := range p.Volumes {
			for name, mountPath := range p.Containers[i].VolumeMounts {
				if name == volume.Name {
					p.Containers[i].VolumeMounts[volume.Path] = mountPath
				}
			}
		}

		err := container.EnsureImage(p.Containers[i].Image)
		if err != nil {
			fmt.Printf("Failed to pull container image: %v\n", err)
			allContainersRunning = false
			break
		}

		volumeBinds := []string{}
		for _, volume := range p.Volumes {
			volumeBinds = append(volumeBinds, fmt.Sprintf("%s:%s", volume.Path, volume.Name))
		}

		err, ID := container.CreateContainer(&p.Containers[i], pauseMode, volumeBinds)
		if err != nil {
			fmt.Printf("Failed to create container: %v\n", err)
			allContainersRunning = false
			break
		}

		err = container.StartContainer(ID)
		if err != nil {
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
		p.PodStatus = Succeeded
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
	if p.PodStatus == Pending {
		fmt.Printf("Pod %s is pending.\n", p.Name)
	} else {
		duration := time.Since(p.StartTime)
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
				err, ID := container.CreateContainer(&p.Containers[i], "container:"+p.PauseContainer.Name, nil)
				if err != nil {
					fmt.Printf("Failed to restart container: %v\n", err)
					p.PodStatus = Failed
				} else {
					err = container.StartContainer(ID)
					if err != nil {
						fmt.Printf("Failed to start container: %v\n", err)
						p.PodStatus = Failed
					} else {
						p.PodStatus = Running
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
