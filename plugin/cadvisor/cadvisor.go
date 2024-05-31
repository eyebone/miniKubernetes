package cadvisor

import (
	"fmt"
	"github.com/google/cadvisor/client"
	_ "github.com/google/cadvisor/info/v1"
	info "github.com/google/cadvisor/info/v1"
	"log"
)

func main() {
	// 连接到cAdvisor API
	cadvisorUrl := "http://localhost:8080/"
	client, err := client.NewClient(cadvisorUrl)
	if err != nil {
		log.Fatalf("Error creating cAdvisor client: %v", err)
	}

	// 获取所有容器的信息
	request := info.ContainerInfoRequest{NumStats: 1}
	containers, err := client.AllDockerContainers(&request)
	if err != nil {
		log.Fatalf("Error getting container info: %v", err)
	}

	// 打印每个容器的信息
	for _, container := range containers {
		fmt.Printf("Container Name: %s\n", container.Name)
		fmt.Printf("Container ID: %s\n", container.Id)
		fmt.Printf("CPU Usage: %v\n", container.Stats[0].Cpu.Usage.Total)
		fmt.Printf("Memory Usage: %v\n", container.Stats[0].Memory.Usage)
		fmt.Printf("Network Usage: %v\n", container.Stats[0].Network.Interfaces)
		fmt.Println("=====================================")
	}
}
