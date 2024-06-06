package test_pod

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"new_k8s/pkg/test_pod/random"
	"new_k8s/tools/etcd/etcd"
	"sync"
)

type PodManager struct {
	mu      sync.RWMutex
	pods    map[string]*Pod
	podUid  map[string]*Pod
	etcdCli etcd.MyEtcdClient
}

func NewPodManager(etcdCli etcd.MyEtcdClient) *PodManager {
	return &PodManager{
		pods:    make(map[string]*Pod),
		podUid:  make(map[string]*Pod),
		etcdCli: etcdCli,
	}
}

/*
*  NOTE: this function returns the one and only Uid of pod
*  Note: in this function , pm dealt with the mapping of pod name and pod; podUid and pod
 */

func (pm *PodManager) CreateNewPod(ctx context.Context, cli *etcd.MyEtcdClient, filename string) string {
	p, err := NewPod(filename, *cli)
	if err != nil {
		log.Fatalf("Failed to create pod: %v", err)
	}
	if p.Configs.Metadata.Uid == "" {
		p.Configs.Metadata.Uid = random.GenerateXid()
	}

	//pm.AddPod(&p)

	// 状态标记为Pending
	p.PodPhase = Pending
	//写入Pod元数据信息到etcd
	podKey := fmt.Sprintf("pods/%s", p.Configs.Metadata.Uid)
	podData, err := json.Marshal(p)
	if err != nil {
		log.Fatalf("Failed to marshal pod data: %v", err)
	}
	if err := cli.Put(podKey, string(podData)); err != nil {
		log.Fatalf("Failed to write pod data to etcd: %v", err)
	}

	return p.Configs.Metadata.Uid
}

// 函数里调用了podfunc中的start方法
/**
* NOTE:
 */

func (pm *PodManager) StartPod(ctx context.Context, cli *etcd.MyEtcdClient, podUid string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// 从etcd中读取Pod信息
	podKey := fmt.Sprintf("pods/%s", podUid)

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

	fmt.Printf("========== Pod %s started successfully! =========\n", p.Configs.Metadata.Name)

	// 更新etcd中的Pod信息
	// TODO: 之后要修改！
	updatedPodData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal updated pod data: %v\n", err)
	}
	// NOTE: 此处的podkey已经修改为 `pods/<pod-uid>`

	podKey = "pods/" + p.Configs.Metadata.Uid
	if err := pm.etcdCli.Put(podKey, string(updatedPodData)); err != nil {
		return fmt.Errorf("failed to update pod data in etcd: %v\n", err)
	}
	podMapKey := "pods"
	err = pm.etcdCli.Put(podMapKey, podUid)
	if err != nil {
		return fmt.Errorf("put pod into etcd pod map failed. %v\n", err)
	}
	// NOTE: 记得更新这两个映射
	//pm.AddPod(&p)
	return nil
}

// TODO: 修改函数逻辑！输入的参数应为podname
func (pm *PodManager) StopPod(ctx context.Context, cli *client.Client, podUid string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 从 Etcd 获取最新的 Pod 信息
	podKey := fmt.Sprintf("pods/%s", podUid)
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

// TODO: 记得更新 pm 中的映射
func (pm *PodManager) RemovePod(ctx context.Context, cli *client.Client, etcdClient *etcd.MyEtcdClient, podUid string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// 从 Etcd 获取最新的 Pod 信息
	podKey := fmt.Sprintf("pods/%s", podUid)
	podData, err := pm.etcdCli.Get(podKey)
	if err != nil {
		return fmt.Errorf("podManager removePod func: failed to get pod data from etcd: %v\n", err)
	}

	var p Pod
	if err := json.Unmarshal([]byte(podData), &p); err != nil {
		return fmt.Errorf("failed to unmarshal pod data: %v", err)
	}
	p.Remove(ctx, cli)

	err = etcdClient.Delete(podKey)
	if err != nil {
		return fmt.Errorf("podManager removePod func: failed to remove pod: %v", err)
	}

	fmt.Printf("Pod %s is removed\n", podUid)

	delete(pm.pods, p.Configs.Metadata.Name)
	delete(pm.podUid, podUid)

	return nil
}

func (pm *PodManager) GetPod(podUids []string, etcdClient *etcd.MyEtcdClient) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	fmt.Printf("POD_ID            POD_NAME         POD_PHASE       RUN_TIME         POD_IP \n")

	// 从 Etcd 获取最新的 Pod 信息
	for i := range len(podUids) {
		podKey := fmt.Sprintf("pods/%s", podUids[i])
		podData, err := pm.etcdCli.Get(podKey)
		if err != nil {
			return fmt.Errorf("podManager displayPod func: failed to get pod data from etcd: %v\n", err)
		}
		var p Pod
		if err := json.Unmarshal([]byte(podData), &p); err != nil {
			return fmt.Errorf("failed to unmarshal pod data: %v", err)
		}
		if p.PodPhase != Pending {
			p.DisplayStatus(pm)
		}

	}

	return nil
}

// TODO: 修改此函数逻辑
func (pm *PodManager) DescribePod(podID string, etcdClient *etcd.MyEtcdClient) {
	//for po := range pm.pods {
	//	fmt.Printf("pod name: %s\n", po)
	//}
	//fmt.Printf("pod name mapping length: %v", len(pm.pods))
	//pod, ok := pm.GetPodByName(podName)
	//if !ok {
	//	log.Fatalf("podName does not exists. \n")
	//}
	podKey := fmt.Sprintf("pods/%s", podID)
	podData, err := pm.etcdCli.Get(podKey)
	if err != nil {
		fmt.Errorf("podManager describe func: failed to get pod data from etcd: %v\n", err)
	}

	var p Pod
	if err := json.Unmarshal([]byte(podData), &p); err != nil {
		fmt.Errorf("failed to unmarshal pod data: %v", err)
	}

	p.DescribePod()
}

func (pm *PodManager) GetPodByName(podName string) (*Pod, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pod, ok := pm.pods[podName]
	fmt.Printf("get pod by name: %v\n", pod)
	return pod, ok
}

func (pm *PodManager) GetPodByuID(podUid string) (*Pod, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pod, ok := pm.podUid[podUid]
	return pod, ok
}

func (pm *PodManager) AddPod(pod *Pod) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	// 插入到 pods 映射
	pm.pods[pod.Configs.Metadata.Name] = pod
	fmt.Printf("add pod to pod name map: %v\n", pm.pods[pod.Configs.Metadata.Name])
	// 插入到 podUid 映射
	pm.podUid[pod.Configs.Metadata.Uid] = pod
	fmt.Printf("add pod to pod uid map: %v\n", pm.podUid[pod.Configs.Metadata.Uid])

}
