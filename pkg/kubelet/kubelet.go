package kubelet

type Kubelet struct {
	HostName string `json "hostName" yaml:"hostName"`
	HostIP   string `json "hostIP" yaml:"hostIP"`
}

func main() {

}
