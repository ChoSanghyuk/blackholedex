package configs

import (
	blackholedex "blackholego"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the entire configuration structure from config.yml
type Config struct {
	RPC              string                            `yaml:"rpc"`
	ContractClient   map[string]ContractClientYAMLData `yaml:"contract_client"`
	StrategyYAMLData StrategyYAMLData                  `yaml:"strategy"`
}

// ContractClientYAMLData represents a single contract configuration from YAML
type ContractClientYAMLData struct {
	Address string `yaml:"address"`
	ABI     string `yaml:"abi"`
}

type StrategyYAMLData struct {
	MonitoringInterval      int     `yaml:"monitoringIntervalSec"`
	StabilityThreshold      float64 `yaml:"stabilityThreshold"`
	StabilityIntervals      int     `yaml:"stabilityIntervals"`
	RangeWidth              int     `yaml:"rangeWidth"`
	SlippagePct             int     `yaml:"slippagePct"`
	CircuitBreakerWindow    int     `yaml:"circuitBreakerWindowMin"`
	CircuitBreakerThreshold int     `yaml:"circuitBreakerThreshold"`
	InitPhase               int     `yaml:"initPhase"`
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

func (c *Config) ToBlackholeConfigs(pk string) *blackholedex.BlackholeConfig {
	var configs []blackholedex.ContractClientConfig

	for _, data := range c.ContractClient {
		configs = append(configs, blackholedex.ContractClientConfig{
			Address: data.Address,
			Abipath: data.ABI,
		})
	}
	return blackholedex.NewBlackholeConfig(
		c.RPC,
		pk,
		nil, // todo. 필요시 config.yaml에서 별도 설정.
		configs,
	)
}

func (c *Config) ToStrategyConfig() *blackholedex.StrategyConfig {
	return &blackholedex.StrategyConfig{
		MonitoringInterval:      time.Duration(c.StrategyYAMLData.MonitoringInterval) * time.Second,
		StabilityThreshold:      c.StrategyYAMLData.StabilityThreshold,
		StabilityIntervals:      c.StrategyYAMLData.StabilityIntervals,
		RangeWidth:              c.StrategyYAMLData.RangeWidth,
		SlippagePct:             c.StrategyYAMLData.SlippagePct,
		CircuitBreakerWindow:    time.Duration(c.StrategyYAMLData.CircuitBreakerWindow) * time.Minute,
		CircuitBreakerThreshold: c.StrategyYAMLData.CircuitBreakerThreshold,
		// InitPhase:               blackholedex.StrategyPhase(c.StrategyYAMLData.InitPhase),
	}
}

// // ToContractClientConfigs converts the Config struct into a slice of ContractClientConfig
// // This method returns the format expected by blackholedex.NewBlackhole()
// func (c *Config) ToContractClientConfigs() []blackholedex.ContractClientConfig {
// 	var configs []blackholedex.ContractClientConfig

// 	for _, data := range c.ContractClient {
// 		configs = append(configs, blackholedex.ContractClientConfig{
// 			Address: data.Address,
// 			Abipath: data.ABI,
// 		})
// 	}

// 	return configs
// }
