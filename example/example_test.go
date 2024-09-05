package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/ploynomail/go-netplan-types/v2"
	yamlnillable "github.com/ploynomail/go-yaml-nillable"
	"github.com/ploynomail/gonetplan"
)

type Logger struct{}

func (l *Logger) Debugf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf("Debugf: "+format, v...)
}
func (l *Logger) Errorf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf("Errorf: "+format, v...)
}

func TestExmaple(t *testing.T) {
	np := gonetplan.NewNetPlan("/etc/netplan", 101, "uva-agent.yaml", &Logger{})
	nc, err := np.GetNetConfig()
	if err != nil {
		t.Error(err)
	}
	if nc.Network != nil && nc.Network.Ethernets != nil {
		if e, ok := nc.Network.Ethernets["enp0s8"]; ok {
			e.DHCP4 = yamlnillable.BoolOf(false)
		}

	}
	nc.Network.Bonds = map[string]*netplan.Bond{}
	nc.Network.Bonds["bond0"] = &netplan.Bond{
		Interfaces: []string{"enp0s9", "enp0s10"},
		Parameters: &netplan.BondParameters{
			Mode: netplan.IEEE8023adBondMode(),
		},
	}
	if err := np.SaveNetConfig(nc); err != nil {
		t.Error(err)
	}
	nc.Network.Bonds = map[string]*netplan.Bond{}
	if err := np.SaveNetConfig(nc); err != nil {
		t.Error(err)
	}
}
