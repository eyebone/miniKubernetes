package test_pod

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"new_k8s/etcd"
	"sync"
)

type PodManager struct {
	mu      sync.RWMutex
	pods    map[string]*Pod
	etcdCli etcd.MyEtcdClient
}

func NewPodManager(etcdCli etcd.MyEtcdClient) *PodManager {
	return &PodManager{
		pods:    make(map[string]*Pod),
		etcdCli: etcdCli,
	}
}

func (pm *PodManager) AddPod(p *Pod) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// 将Pod信息序列化为JSON字符串
	podData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal pod data: %v", err)
	}
	// 写入etcd
	if err := pm.etcdCli.Put(p.Configs.Metadata.Name, string(podData)); err != nil {
		return fmt.Errorf("failed to put pod data to etcd: %v", err)
	}
	pm.pods[p.Configs.Metadata.Name] = p
	return nil
}

// 函数里调用了podfunc中的start方法

func (pm *PodManager) StartPod(ctx context.Context, cli *etcd.MyEtcdClient, podName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// 从etcd中读取Pod信息
	podKey := fmt.Sprintf("pods/%s", podName)
	podData, err := pm.etcdCli.Get(podKey)
	if err != nil {
		return fmt.Errorf("failed to get pod data from etcd: %v\n", err)
	}

	var p Pod
	// 反序列化从etcd中获得的pod元数据信息
	if err := json.Unmarshal([]byte(podData), &p); err != nil {
		return fmt.Errorf("failed to unmarshal pod data: %v", err)
	} else {
		fmt.Printf("pod manager start pod: get Pod data from etcd: %+v\n", podData)
	}
	/**
	   调用pod中的start方法！
	  * 这里的cli是etcdclient！
	*/
	p.Start(ctx, cli)
	if p.PodPhase != Running {
		fmt.Printf("Failed to start pod: %v\n", p.PodPhase)
		return fmt.Errorf("failed to start pod: %v", p.PodPhase)
	}

	fmt.Printf("Pod %s started successfully\n", p.Configs.Metadata.Name)

	// 更新etcd中的Pod信息
	updatedPodData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal updated pod data: %v", err)
	}
	podKey = "pods/" + p.Configs.Metadata.Name
	if err := pm.etcdCli.Put(podKey, string(updatedPodData)); err != nil {
		return fmt.Errorf("failed to update pod data in etcd: %v", err)
	}
	pm.pods[p.Configs.Metadata.Name] = &p
	return nil
}

func (pm *PodManager) StopPod(ctx context.Context, cli *client.Client, podName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 从 Etcd 获取最新的 Pod 信息
	podKey := fmt.Sprintf("pods/%s", podName)
	podData, err := pm.etcdCli.Get(podKey)
	if err != nil {
		return fmt.Errorf("podManager stopPod func: failed to get pod data from etcd: %v\n", err)
	}

	var p Pod
	if err := json.Unmarshal([]byte(podData), &p); err != nil {
		return fmt.Errorf("failed to unmarshal pod data: %v", err)
	}

	p.Stop(ctx, cli)
	if p.PodPhase == Failed {
		fmt.Printf("pod can't be stopped, pod phase is: %v\n", p.PodPhase)
	} else if p.PodPhase == Succeeded {
		fmt.Printf("pod stopped successfully, pod phase is: %v\n", p.PodPhase)
	}

	// 更新 Etcd 中的 Pod 信息
	updatedPodData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal updated pod data: %v", err)
	}
	if err := pm.etcdCli.Put(podKey, string(updatedPodData)); err != nil {
		return fmt.Errorf("failed to update pod data in etcd: %v", err)
	}

	return nil
}

func (pm *PodManager) RemovePod(ctx context.Context, cli *client.Client, podName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// 从 Etcd 获取最新的 Pod 信息
	podKey := fmt.Sprintf("pods/%s", podName)
	podData, err := pm.etcdCli.Get(podKey)
	if err != nil {
		return fmt.Errorf("podManager removePod func: failed to get pod data from etcd: %v\n", err)
	}

	var p Pod
	if err := json.Unmarshal([]byte(podData), &p); err != nil {
		return fmt.Errorf("failed to unmarshal pod data: %v", err)
	}
	err = p.Remove(ctx, cli)
	if err != nil {
		return fmt.Errorf("podManager removePod func: failed to remove pod: %v", err)
	}
	fmt.Printf("Pod %s is removed\n", podName)
	delete(pm.pods, podName)
	return nil
}

func (pm *PodManager) DisplayPodStatus(podName string) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	// 从 Etcd 获取最新的 Pod 信息
	podKey := fmt.Sprintf("pods/%s", podName)
	podData, err := pm.etcdCli.Get(podKey)
	if err != nil {
		return fmt.Errorf("podManager displayPod func: failed to get pod data from etcd: %v\n", err)
	}

	var p Pod
	if err := json.Unmarshal([]byte(podData), &p); err != nil {
		return fmt.Errorf("failed to unmarshal pod data: %v", err)
	}
	p.DisplayStatus()
	return nil
}
