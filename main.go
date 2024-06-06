// package main
//
// import (
//
//	"context"
//	"fmt"
//	"log"
//	"os"
//	"path/filepath"
//
//	"github.com/docker/docker/client"
//	"new_k8s/pkg/pod"
//
// )
//
//	func main() {

//

//
//		// 启动监控所有Pod的协程
//		go podManager.Monitor(ctx, cli)
//
//		// 加载现有的Pod配置
//		loadExistingPods(podManager, cli)
//
//		switch command {
//		case "start":
//			absPath, err := filepath.Abs(arg)
//			if err != nil {
//				log.Fatalf("Error getting absolute path: %v", err)
//			}
//			fmt.Printf("=====starting to create a new pod\n")
//			p, err := pod.NewPod(absPath)
//			if err != nil {
//				log.Fatalf("Error creating pod: %v", err)
//			}
//			podManager.StartPod(ctx, cli, p)
//		case "stop":
//			if err := podManager.StopPod(ctx, cli, arg); err != nil {
//				log.Fatalf("Error stopping pod: %v", err)
//			}
//		case "remove":
//			if err := podManager.RemovePod(ctx, cli, arg); err != nil {
//				log.Fatalf("Error removing pod: %v", err)
//			}
//		case "get":
//			if err := podManager.DisplayPodStatus(arg); err != nil {
//				log.Fatalf("Error getting pod status: %v", err)
//			}
//		case "describe":
//			if err := podManager.DisplayPodStatus(arg); err != nil {
//				log.Fatalf("Error describing pod: %v", err)
//			}
//		default:
//			fmt.Println("Unknown command. Usage: [start | stop | remove | get | describe] <pod yaml file or pod name>")
//		}
//	}
//
// // 从yaml文件加载现有的Pod
//

package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	test_pod2 "new_k8s/pkg/test_pod/test_pod"
	"new_k8s/tools/etcd/etcd"
	"os"
	"os/exec"
	"sync"
)

var once sync.Once
var etcdCmd *exec.Cmd
var cli *clientv3.Client

func main() {
	// 处理命令行输入
	if len(os.Args) < 3 {
		fmt.Println("Usage: [start | stop | remove | get | describe] <pod yaml file or pod name>")
		return
	}
	// etcd 启动
	err := etcd.StartEtcd()
	if err != nil {
		fmt.Printf("error starting etcd: \n")
		log.Fatal(err)
	}
	defer func() {
		if etcdCmd != nil && etcdCmd.Process != nil {
			if err := etcdCmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill etcd process: ", err)
			}
		}
	}()
	// 连接etcd
	cliEtcd, err := etcd.ConnectEtcd()
	if err != nil {
		log.Fatal(err)
	}
	defer cliEtcd.Close()

	//fmt.Println("Connected to etcd successfully!\n")
	// 创建 etcd 客户端
	etcdClient := &etcd.MyEtcdClient{Client: cliEtcd}

	ctx := context.Background()

	// 创建docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	// 解析命令行参数
	command := os.Args[1]
	arg := os.Args[2]
	// 创建podManager
	podManager := test_pod2.NewPodManager(*etcdClient)

	//loadExistingPods(podManager, etcdClient)

	switch command {
	case "start":
		// 获取yaml文件路径
		newFileName := getFilePath(arg)
		// 处理启动pod逻辑
		handleStartCmd(ctx, cli, etcdClient, newFileName, *podManager)

	case "stop":
		fmt.Printf("=====starting to stop pod========\n")
		podManager.StopPod(ctx, cli, arg)
	case "get":
		fmt.Printf("======get pod status:======\n")

		// NOTE: 此处data为 `[]string` 类型，里面为所有pod的Uid
		data := etcd.GetPodPrefixKeys()
		podManager.GetPod(data, etcdClient)
	case "describe":
		podManager.DescribePod(arg, etcdClient)
	case "delete":
		podManager.RemovePod(ctx, cli, etcdClient, arg)
	}

}

func getFilePath(arg string) string {
	// 打印接收到的命令行参数
	//fmt.Printf("Received file parameter: %s\n", arg)
	// 打印当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	//fmt.Printf("Current working directory: %s\n", wd)
	newFileName := wd + arg
	return newFileName
}

func handleStartCmd(ctx context.Context, cli *client.Client, etcdClient *etcd.MyEtcdClient, newFileName string, podManager test_pod2.PodManager) {
	fmt.Printf("=====starting to create a new pod========\n")
	// TODO: 之后要改，这个逻辑应该要在pm中
	podUid := podManager.CreateNewPod(ctx, etcdClient, newFileName)
	// 调用podManager 的startPod方法
	podManager.StartPod(ctx, etcdClient, podUid)
}

//func loadExistingPods(pm *test_pod2.PodManager, cli *etcd.MyEtcdClient) {
//	files, err := filepath.Glob("*.yaml")
//	if err != nil {
//		log.Fatalf("Error loading pod configurations: %v", err)
//	}
//
//	for _, file := range files {
//		p, _, err := test_pod2.NewPod(file, *cli)
//		if err != nil {
//			log.Printf("Error creating pod from file %s: %v", file, err)
//			continue
//		}
//		pm.AddPod(&p)
//	}
//}
