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

	//"flag"
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"new_k8s/pkg/pod/test_pod"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: [start | stop | remove | get | describe] <pod yaml file or pod name>")
		return
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	command := os.Args[1]
	arg := os.Args[2]

	podManager := test_pod.NewPodManager()

	loadExistingPods(podManager, cli)

	switch command {
	case "start":
		// 打印接收到的命令行参数
		fmt.Printf("Received file parameter: %s\n", arg)
		// 打印当前工作目录
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
		fmt.Printf("Current working directory: %s\n", wd)
		// 读取 YAML 文件
		newFileName := wd + arg
		fmt.Printf("=====starting to create a new pod========\n")
		p, err := test_pod.NewPod(newFileName)
		if err != nil {
			log.Fatalf("Failed to create pod: %v", err)
		}

		// 打印解析后的结构体
		fmt.Printf("Parsed Pod from file %s: %+v\n", newFileName, p)
		//fmt.Printf("pod volume path: %s")
		podManager.StartPod(ctx, cli, &p)

	}

	//// 定义一个字符串类型的命令行参数用于传递 YAML 文件名
	//fileName := flag.String("file", "", "YAML file to parse")
	//flag.Parse()
	//
	//// 打印接收到的命令行参数
	//fmt.Printf("Received file parameter: %s\n", *fileName)
	//
	//// 打印当前工作目录
	//wd, err := os.Getwd()
	//if err != nil {
	//	log.Fatalf("Error getting current working directory: %v", err)
	//}
	//fmt.Printf("Current working directory: %s\n", wd)

	// 处理相对路径
	//absFileName, err := filepath.Abs(*fileName)
	//if err != nil {
	//	log.Fatalf("Error converting to absolute path: %v", err)
	//}
	//fmt.Printf("Absolute file path: %s\n", absFileName)
	//
	//// 检查是否提供了文件名
	//if *fileName == "" {
	//	log.Fatalf("Please provide a YAML file using the -file flag")
	//}

	// 读取 YAML 文件
	//newFileName := wd + *fileName + ".yaml"

	// 解析 YAML 文件
	// 调用 newPod 函数并处理结果
	//pod, err := test_pod.NewPod(newFileName)

}

func loadExistingPods(pm *test_pod.PodManager, cli *client.Client) {
	files, err := filepath.Glob("*.yaml")
	if err != nil {
		log.Fatalf("Error loading pod configurations: %v", err)
	}

	for _, file := range files {
		p, err := test_pod.NewPod(file)
		if err != nil {
			log.Printf("Error creating pod from file %s: %v", file, err)
			continue
		}
		pm.AddPod(&p)
	}
}
