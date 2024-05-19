package main

import (
	"context"
	"flag"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"time"
)

// 表示对 etcd 的增删改查操作的接口
/*type EtcdClient interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Update(key, value string) error
}

// 实现 `EtcdClient` 接口的结构体
type MyEtcdClient struct {
	Client *clientv3.Client
}

func (ec *MyEtcdClient) Put(key, value string) error {
	_, err := ec.Client.Put(context.Background(), key, value)
	if err != nil {
		return err
	}
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
	return string(resp.Kvs[0].Value), nil
}

func (ec *MyEtcdClient) Delete(key string) error {
	_, err := ec.Client.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	return nil
}

func (ec *MyEtcdClient) Update(key, value string) error {
	_, err := ec.Client.Put(context.Background(), key, value)
	if err != nil {
		return err
	}
	return nil
}*/

func main() {
	// 命令行参数
	var (
		endpoints   string
		dialTimeout time.Duration
		cmd         string
		key         string
		value       string
	)

	flag.StringVar(&endpoints, "endpoints", "localhost:2379", "etcd endpoints, separated by commas")
	flag.DurationVar(&dialTimeout, "dialTimeout", 5*time.Second, "etcd connection dialtimeout duration")
	flag.StringVar(&cmd, "cmd", "", "command to execute operations in etcd")
	flag.StringVar(&key, "key", "", "key used to operate delete/get/put")
	flag.StringVar(&value, "value", "", "value used to operate put")
	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "etcdclient")
		flag.PrintDefaults()
		fmt.Println("Example:")
		fmt.Println("  ./etcdclient -endpoints 'localhost:2379,localhost:2330,localhost:8997' -dialTimeout 5 -cmd put -key 'k8s' -value 'shit'")
	}

	// 解析命令行参数
	flag.Parse()

	// 连接 etcd 服务器
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
		fmt.Println("conntect to etcd failed:", err)
	}
	defer cli.Close()

	fmt.Println("Connected to etcd successfully")

	// 创建 etcd 客户端
	etcdClient := &MyEtcdClient{Client: cli}

	// 执行命令
	switch cmd {
	case "put":
		if key == "" || value == "" {
			log.Fatal("key and value are required for put command")
		}
		err := etcdClient.Put(key, value)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Put operation completed successfully")
	case "get":
		if key == "" {
			log.Fatal("key is required for get command")
		}
		val, err := etcdClient.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Get: %s\n", val)
	case "delete":
		if key == "" {
			log.Fatal("key is required for delete command")
		}
		err := etcdClient.Delete(key)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Delete operation completed successfully")
	default:
		log.Fatal("invalid command")
	}
}

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
	fmt.Println("Get operation completed successfully")
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

/*
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			// handle error!
			fmt.Printf("connect to etcd failed, err:%v\n", err)
			log.Fatal(err)
			return
		}
		defer cli.Close() // 在函数结束前关闭客户端

		fmt.Println("connected to etcd success")

		// 创建 MyEtcdClient 实例
		myEtcdClient := &MyEtcdClient{Client: cli}

		// 使用该实例进行操作
		key := "cxtkey"
		value := "cxtvalue"

		err = myEtcdClient.Put(key, value)
		fmt.Printf("Putting key '%s' with value '%s'\n", key, value)
		if err != nil {
			log.Fatal(err)
		}

		val, err := myEtcdClient.Get(key)
		fmt.Printf("Getting key '%s'\n", key)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Get: ", val)

		err = myEtcdClient.Delete(key)
		fmt.Printf("Deleting key '%s'\n", key)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("updating key '%s'\n", key)
		err = myEtcdClient.Update(key, "this is lyf")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("updating success")
		fmt.Println("Operation completed successfully")


}
*/
