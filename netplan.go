package gonetplan

import (
	"context"

	"github.com/ploynomail/go-netplan-types/v2"
)

type NetPlan struct {
	npc *NetPlanConfig
	log Logger
}

func NewNetPlan(
	configPath string,
	TargetConfigPriority int,
	TargetConfig string,
	log Logger,
) *NetPlan {
	return &NetPlan{
		npc: NewNetPlanConfig(context.Background(), configPath, TargetConfigPriority, TargetConfig, log),
		log: log,
	}
}

func (np *NetPlan) GetNetConfig() (netplan.Network, error) {
	return np.npc.GetNetplanConfig()
}

func (np *NetPlan) SaveNetConfig(nc netplan.Network) error {
	removeDev := np.FindRemoveVirtualDevices(nc)
	if err := np.RemoveVirtualDevices(removeDev); err != nil {
		np.log.Errorf(context.Background(), "Failed to remove virtual devices: %v", err)
		return err
	}
	if err := np.npc.SaveNetplanConfig(nc); err != nil {
		np.log.Errorf(context.Background(), "Failed to save netplan config: %v", err)
		return err
	}
	if err := ApplyNetplanConfig(); err != nil {
		np.log.Errorf(context.Background(), "Failed to apply netplan config: %v", err)
		return err
	}
	return nil
}

func (np *NetPlan) RemoveVirtualDevices(devName []string) error {
	for _, dev := range devName {
		if err := RemoveVirtualDevices(dev); err != nil {
			np.log.Errorf(context.Background(), "Failed to remove virtual device %s: %v", dev, err)
			return err
		}
	}
	return nil
}

func (np *NetPlan) FindRemoveVirtualDevices(nc netplan.Network) []string {
	devName := []string{}
	for name := range nc.Network.Bonds {
		devName = append(devName, name)
	}
	for name := range nc.Network.Bridges {
		devName = append(devName, name)
	}
	for name := range nc.Network.VLANs {
		devName = append(devName, name)
	}
	exitsDevName := GetBondBridgeVlanNames()

	RemoveDevName := difference(devName, exitsDevName)
	return RemoveDevName
}
