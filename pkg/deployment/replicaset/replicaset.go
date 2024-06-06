package replicaset

type ReplicaSet struct {
	Configs struct {
		kind     string `json:"kind" yaml:"kind"`
		Metadata struct {
			Name string `json:"name" yaml:"name"`
		}
		Spec struct {
			Replicas int32 `json:"replicas" yaml:"replicas"`
			Selector struct {
				matchName string `json:"matchName" yaml:"matchName"`
			}
		}
	}
}

func main() {

}
