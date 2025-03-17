package container

import (
	"net"
)

type Network struct {
	Name      string
	Bridge    string
	IPRange   *net.IPNet
	Gateway   net.IP
	Interface string
}

func NewNetwork(name string) *Network {
	return &Network{
		Name:    name,
		Bridge:  "gontainer0",
		IPRange: &net.IPNet{IP: net.ParseIP("172.17.0.0"), Mask: net.CIDRMask(16, 32)},
		Gateway: net.ParseIP("172.17.0.1"),
	}
}

func (n *Network) Setup() error {
	// Basic network setup implementation
	return nil
}

func (c *Container) SetupNetwork() *Network {
	// Create a basic network configuration
	network := &Network{
		// Default values, can be customized later
		IPRange: &net.IPNet{
			IP:   net.ParseIP("10.0.0.0"),
			Mask: net.CIDRMask(24, 32),
		},
		Gateway: net.ParseIP("10.0.0.1"),
	}
	c.Network = network
	return network
}
