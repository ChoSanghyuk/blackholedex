package blackholedex

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// ContractRegistry manages a map of named contract clients
// Provides lookup by name or address for any contract interaction
// This is a domain-agnostic utility that can be moved to pkg/ if needed in other packages.
type ContractRegistry struct {
	clients map[string]ContractClient
}

// NewContractRegistry creates a registry from contract client map
func NewContractRegistry(clients map[string]ContractClient) *ContractRegistry {
	return &ContractRegistry{
		clients: clients,
	}
}

// Client retrieves a contract client by registered name
func (r *ContractRegistry) Client(name string) (ContractClient, error) {
	c := r.clients[name]
	if c == nil {
		return nil, fmt.Errorf("no mapped client for name: %s", name)
	}
	return c, nil
}

// ClientByAddress finds a contract client by its contract address
func (r *ContractRegistry) ClientByAddress(address string) (ContractClient, error) {
	for _, c := range r.clients {
		if strings.EqualFold(address, c.ContractAddress().Hex()) {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no mapped client for address: %s", address)
}

// GetAddress retrieves the contract address for a given contract name
func (r *ContractRegistry) GetAddress(name string) (common.Address, error) {
	client, err := r.Client(name)
	if err != nil {
		return common.Address{}, err
	}
	return *client.ContractAddress(), nil
}
