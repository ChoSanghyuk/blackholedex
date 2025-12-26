package configs

import (
	blackholedex "blackholego"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the entire configuration structure from config.yml
type Config struct {
	RPC            string                            `yaml:"rpc"`
	ContractClient map[string]ContractClientYAMLData `yaml:"contract_client"`
}

// ContractClientYAMLData represents a single contract configuration from YAML
type ContractClientYAMLData struct {
	Address string `yaml:"address"`
	ABI     string `yaml:"abi"`
}

// ContractClientConfig represents the configuration needed by blackholedex.Blackhole
// This matches the struct defined in blackhole.go:43-46
type ContractClientConfig struct {
	Address string
	ABIPath string
}

// LoadConfig reads and parses config.yml into a Config struct
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return &config, nil
}

// ToContractClientConfigs converts the Config struct into a slice of ContractClientConfig
// This method returns the format expected by blackholedex.NewBlackhole()
func (c *Config) ToContractClientConfigs() []blackholedex.ContractClientConfig {
	var configs []blackholedex.ContractClientConfig

	for _, data := range c.ContractClient {
		configs = append(configs, blackholedex.ContractClientConfig{
			Address: data.Address,
			Abipath: data.ABI,
		})
	}

	return configs
}
