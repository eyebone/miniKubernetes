package test_pod

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"new_k8s/pkg/container"
	"time"
)

func (pod *Pod) Start(ctx context.Context, cli *client.Client) {
	pod.StartTime = time.Now()
	pod.PauseContainer = container.NewContainer()
	pod.PauseContainer.Image = PauseImage

	pauseID, pauseName, err := createPause(pod)
	pod.PauseContainer.Name = pauseName
	if err != nil {
		fmt.Printf("create pause container err:%v", err)
	}
	fmt.Printf("pause container create succeed! containerID: ", pauseID, "containerName: ", pauseName)

	pauseMode := "container:" + pauseID
	allContainersRunning := true

	for i := range pod.configs.Spec.Containers {
		for _, volume := range pod.configs.Spec.Volumes {
			volumeMount := pod.configs.Spec.Containers[i].VolumeMount
			if volumeMount.Name == volume.Name {
				pod.configs.Spec.Containers[i].VolumeMount.MountPath = volume.HostPath.Path
			}
		}

		err := container.EnsureImage(pod.configs.Spec.Containers[i].Image)
		if err != nil {
			fmt.Printf("Failed to pull container image: %v\n", err)
			allContainersRunning = false
			continue
		}

		volumeBinds := []string{}
		for _, volume := range pod.configs.Spec.Volumes {
			volumeBinds = append(volumeBinds, fmt.Sprintf("%v:%v", volume.Name, volume.HostPath.Path))

		}
		cont := container.NewContainer()
		cont.Name = pod.configs.Spec.Containers[i].Name
		cont.Image = pod.configs.Spec.Containers[i].Image
		cont.Ports = pod.configs.Spec.Containers[i].Ports

		ID, contName, err := container.CreateContainer(&pod.configs.Spec.Containers[i], pauseMode, volumeBinds)
		fmt.Printf("hello.im lujyifan.\n")

		if err != nil {
			fmt.Printf("Failed to create container: %v\n", err)
			allContainersRunning = false
			break
		}

		err = container.StartContainer(ID)
		if err != nil {
			fmt.Printf("Failed to start container: %s:%v\n", contName, err)
			allContainersRunning = false
			break
		} else {
			fmt.Printf("********* start container succeed! containerName: %s************\n", contName)
		}
	}

	if allContainersRunning {
		pod.PodPhase = Running
	} else {
		pod.PodPhase = Failed
	}

}

//
//func (pod *Pod) Stop(ctx context.Context, cli *client.Client) {
//	allContainerStooped := true
//
//
//}
