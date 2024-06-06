package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// 定义一个 etcd 客户端接口
type EtcdClient interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

// 实现 EtcdClient 接口的结构体
type MyEtcdClient struct {
	Client *clientv3.Client
}

func (ec *MyEtcdClient) Put(key, value string) error {
	_, err := ec.Client.Put(context.Background(), key, value)
	if err != nil {
		return err
	}
	//fmt.Println("Put operation completed successfully", key, value)
	return nil
}

func (ec *MyEtcdClient) Get(key string) (string, error) {
	resp, err := ec.Client.Get(context.Background(), key)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found")
	}
	//fmt.Println("Get operation completed successfully!\n")
	return string(resp.Kvs[0].Value), nil
}

func (ec *MyEtcdClient) Delete(key string) error {
	_, err := ec.Client.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	//fmt.Println("Delete operation completed successfully")
	return nil
}

var once sync.Once
var etcdCmd *exec.Cmd
var cli *clientv3.Client

func StartEtcd() error {
	var err error
	once.Do(func() {
		// 使用 exec.Command 启动一个新的后台进程运行 etcd
		etcdCmd = exec.Command("nohup", "etcd")
		etcdCmd.Stdout, err = os.OpenFile("etcd_stdout.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
		etcdCmd.Stderr = etcdCmd.Stdout

		err = etcdCmd.Start()
		if err != nil {
			log.Fatalf("failed to start etcd: %v", err)
		}

		// 等待 etcd 启动
		time.Sleep(2 * time.Second)

		//fmt.Println("********** etcd started in background **********\n")
	})
	return err
}

func GetPodPrefixKeys() []string {
	cmd := exec.Command("etcdctl", "get", "pods", "--prefix")
	data, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to call Output(): %v", err)
	}

	output := string(data)
	lines := strings.Split(output, "\n")

	var podNames []string
	for _, line := range lines {
		if strings.Contains(line, "pods/") && !strings.Contains(line, "pause") {
			parts := strings.Split(line, "/")
			if len(parts) > 1 {
				podNames = append(podNames, parts[1])
			}
		}
	}
	//fmt.Printf("podnames: ", podNames)
	return podNames
}

func ConnectEtcd() (*clientv3.Client, error) {
	var err error
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{"127.0.0.1:2379"},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("connect to etcd failed: %v", err)
		}
	}
	return cli, nil
}
