package configs

import (
	"fmt"
	"os"
	"time"

	blackholedex "github.com/ChoSanghyuk/blackholedex"
	"github.com/ChoSanghyuk/blackholedex/pkg/types"

	"gopkg.in/yaml.v3"
)

// Config represents the entire configuration structure from config.yml
type Config struct {
	RPC              string                `yaml:"rpc"`
	ActivePool       string                `yaml:"active_pool"`
	ContractClient   ContractClientSection `yaml:"contract_client"`
	StrategyYAMLData StrategyYAMLData      `yaml:"strategy"`
}

// ContractClientSection represents the contract_client section with common and pool-specific configs
type ContractClientSection struct {
	Common map[string]ContractClientYAMLData `yaml:"common"`
	CL200  map[string]ContractClientYAMLData `yaml:"cl200"`
	CL1    map[string]ContractClientYAMLData `yaml:"cl1"`
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

	// Add common contracts
	for name, data := range c.ContractClient.Common {
		configs = append(configs, blackholedex.ContractClientConfig{
			Name:    name,
			Address: data.Address,
			Abipath: data.ABI,
		})
	}

	// Add pool-specific contracts based on active_pool
	var poolContracts map[string]ContractClientYAMLData
	switch c.ActivePool {
	case "cl1":
		poolContracts = c.ContractClient.CL1
	case "cl200":
		poolContracts = c.ContractClient.CL200
	}

	for name, data := range poolContracts {
		configs = append(configs, blackholedex.ContractClientConfig{
			Name:    name,
			Address: data.Address,
			Abipath: data.ABI,
		})
	}

	var pool types.PoolType
	switch c.ActivePool {
	case "cl1":
		pool = types.CL1
	case "cl200":
		pool = types.CL200
	default:
		pool = types.CL200 // default to CL200 if unknown
	}

	return blackholedex.NewBlackholeConfig(
		c.RPC,
		pk,
		nil, // todo. 필요시 config.yaml에서 별도 설정.
		pool,
		configs,
	)
}

func (c *Config) ToStrategyConfig() *types.StrategyConfig {
	return &types.StrategyConfig{
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
