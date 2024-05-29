package container

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type PortMapping struct {
	ContainerPort int
	HostPort      int
}

type Container struct {
	Name         string
	Image        string
	Command      []string
	CmdArgs      []string
	CPU          int // In milli cores, 1000 milli cores = 1 core
	Memory       int // In MB
	Status       string
	PortMappings []PortMapping
	VolumeMounts map[string]string
}

func NewContainer() *Container {
	return &Container{
		CPU:          1000,
		Memory:       512,
		VolumeMounts: make(map[string]string),
	}
}

func (c *Container) PullImage(ctx context.Context, cli *client.Client) error {
	out, err := cli.ImagePull(ctx, c.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	// Print image pull output
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	fmt.Println(buf.String())
	return nil
}

func (c *Container) Start(ctx context.Context, cli *client.Client) error {
	config := &container.Config{
		Image: c.Image,
		Cmd:   append(c.Command, c.CmdArgs...),
	}
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			CPUShares: int64(c.CPU),
			Memory:    int64(c.Memory * 1024 * 1024),
		},
	}
	netConfig := &network.NetworkingConfig{}
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, netConfig, nil, c.Name)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}
	c.Status = "running"
	return nil
}

func (c *Container) Stop(ctx context.Context, cli *client.Client) error {
	timeout := 10 // Timeout in seconds
	if err := cli.ContainerStop(ctx, c.Name, container.StopOptions{Timeout: &timeout}); err != nil {
		return err
	}
	c.Status = "stopped"
	return nil
}

func (c *Container) Remove(ctx context.Context, cli *client.Client) error {
	if err := cli.ContainerRemove(ctx, c.Name, container.RemoveOptions{}); err != nil {
		return err
	}
	c.Status = "removed"
	return nil
}

func (c *Container) IsRunning(ctx context.Context, cli *client.Client) (bool, error) {
	json, err := cli.ContainerInspect(ctx, c.Name)
	if err != nil {
		return false, err
	}
	return json.State.Running, nil
}
