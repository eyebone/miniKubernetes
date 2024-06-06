package flannel

import (
	"encoding/json"
	"fmt"
)

type FlannelConfig struct {
	Network   string        `json:"Network"`
	SubnetLen int           `json:"SubnetLen"`
	SubnetMin string        `json:"SubnetMin"`
	SubnetMax string        `json:"SubnetMax"`
	Backend   BackendConfig `json:"Backend"`
}
type BackendConfig struct {
	Type string `json:"Type"`
}

func MyFlannelMarshal(respVal []byte) (FlannelConfig, error) {
	var config FlannelConfig
	err := json.Unmarshal(respVal, &config)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return config, err
	}

	fmt.Println("Network:", config.Network)
	return config, nil
}
