package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"os"
	"os/exec"
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
	fmt.Println("Put operation completed successfully", key, value)
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
	fmt.Println("Get operation completed successfully!\n")
	return string(resp.Kvs[0].Value), nil
}

func (ec *MyEtcdClient) Delete(key string) error {
	_, err := ec.Client.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	fmt.Println("Delete operation completed successfully")
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

		fmt.Println("********** etcd started in background **********\n")
	})
	return err
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

//func main() {
// 命令行参数
//var (
//	endpoints   string
//	dialTimeout time.Duration
//	cmd         string
//	key         string
//	value       string
//)
//
//flag.StringVar(&endpoints, "endpoints", "localhost:2379", "etcd endpoints, separated by commas")
//flag.DurationVar(&dialTimeout, "dialTimeout", 5*time.Second, "etcd connection dialtimeout duration")
//flag.StringVar(&cmd, "cmd", "", "command to execute operations in etcd")
//flag.StringVar(&key, "key", "", "key used to operate delete/get/put")
//flag.StringVar(&value, "value", "", "value used to operate put")
//// 自定义帮助信息
//flag.Usage = func() {
//	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "etcdclient")
//	flag.PrintDefaults()
//	fmt.Println("Example:")
//	fmt.Println("  ./etcdclient -endpoints 'localhost:2379,localhost:2330,localhost:8997' -dialTimeout 5 -cmd put -key 'k8s' -value 'value'")
//}
//
//// 解析命令行参数
//flag.Parse()

//	// 启动 etcd
//	err := startEtcd()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer func() {
//		if etcdCmd != nil && etcdCmd.Process != nil {
//			if err := etcdCmd.Process.Kill(); err != nil {
//				log.Fatal("failed to kill etcd process: ", err)
//			}
//		}
//	}()
//
//	// 连接 etcd 服务器
//	cli, err = connectEtcd()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer cli.Close()
//
//	fmt.Println("Connected to etcd successfully")
//
//	// 创建 etcd 客户端
//	etcdClient := &MyEtcdClient{Client: cli}
//
//	// 执行命令
//	switch cmd {
//	case "put":
//		if key == "" || value == "" {
//			log.Fatal("key and value are required for put command")
//		}
//		err := etcdClient.Put(key, value)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println("Put operation completed successfully")
//	case "get":
//		if key == "" {
//			log.Fatal("key is required for get command")
//		}
//		val, err := etcdClient.Get(key)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("Get: %s\n", val)
//	case "delete":
//		if key == "" {
//			log.Fatal("key is required for delete command")
//		}
//		err := etcdClient.Delete(key)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println("Delete operation completed successfully")
//	default:
//		log.Fatal("invalid command")
//	}
//}
