package gonetplan

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/ploynomail/go-netplan-types/v2"
	yamlnillable "github.com/ploynomail/go-yaml-nillable"
	"gopkg.in/yaml.v2"
)

type NetPlanConfig struct {
	ConfigPath   string `json:"config_path"`
	TargetConfig string `json:"target_config"` // The name of the file that this module directly manages
	logger       Logger
	ctx          context.Context
}

func NewNetPlanConfig(
	ctx context.Context,
	configPath string,
	TargetConfigPriority int,
	TargetConfig string,
	log Logger,
) *NetPlanConfig {
	targetConfig := strconv.Itoa(TargetConfigPriority) + "-" + TargetConfig
	return &NetPlanConfig{
		ConfigPath:   configPath,
		TargetConfig: targetConfig,
		logger:       log,
		ctx:          ctx,
	}
}

func (np *NetPlanConfig) SaveNetplanConfig(nc netplan.Network) error {
	o, err := yaml.Marshal(nc)
	if err != nil {
		np.logger.Errorf(np.ctx, "Failed to marshal netplan config: %v", err)
		return err
	}
	configPath := filepath.Join(np.ConfigPath, np.TargetConfig)
	if err := os.WriteFile(configPath, o, 0600); err != nil {
		np.logger.Errorf(np.ctx, "Failed to write netplan config: %v", err)
		return err
	}
	return nil
}

func (np *NetPlanConfig) GetNetplanConfig() (netplan.Network, error) {
	configPath := filepath.Join(np.ConfigPath, np.TargetConfig)
	if _, err := os.Stat(np.ConfigPath); os.IsNotExist(err) {
		np.logger.Errorf(np.ctx, "Netplan config not found: %v", err)
		return netplan.Network{}, err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := np.createTargetConfig(); err != nil {
			np.logger.Errorf(np.ctx, "Failed to create target config: %v", err)
			return netplan.Network{}, err
		}
	}
	netplanConfig, err := os.ReadFile(configPath)
	if err != nil {
		np.logger.Errorf(np.ctx, "Failed to read netplan config: %v", err)
		return netplan.Network{}, err
	}
	if len(netplanConfig) == 0 {
		if err := np.createTargetConfig(); err != nil {
			np.logger.Errorf(np.ctx, "Failed to create target config: %v", err)
			return netplan.Network{}, err
		}
	}
	var nc netplan.Network
	if err := yaml.Unmarshal(netplanConfig, &nc); err != nil {
		np.logger.Errorf(np.ctx, "Failed to unmarshal netplan config: %v", err)
		return netplan.Network{}, err
	}
	return nc, nil
}

func (np *NetPlanConfig) visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			np.logger.Errorf(np.ctx, "Failed to walk path: %v", err)
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			np.logger.Debugf(np.ctx, "Found file: %v", path)
			*files = append(*files, path)
		}
		return nil
	}
}

// 查找配置中不存在的物理接口
func (np *NetPlanConfig) FindMissingPhysicalInterfaces(nc netplan.Network) ([]string, error) {
	links, err := GetEthernetLinks()
	if err != nil {
		np.logger.Errorf(np.ctx, "Failed to get ethernet links: %v", err)
		return nil, err
	}
	missing := []string{}
	for _, link := range links {
		found := false
		for name, _ := range nc.Network.Ethernets {
			if link.Attrs().Name == name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, link.Attrs().Name)
		}
	}
	return missing, nil
}

func (np *NetPlanConfig) createTargetConfig() error {
	configPath := filepath.Join(np.ConfigPath, np.TargetConfig)
	var files []string
	err := filepath.Walk(np.ConfigPath, np.visit(&files))
	if err != nil {
		np.logger.Errorf(np.ctx, "Failed to walk path: %v", err)
		return err
	}
	files = OrderFiles(files)
	var nc netplan.Network = netplan.Network{}
	for _, file := range files {
		netplanConfig, err := os.ReadFile(file)
		if err != nil {
			np.logger.Errorf(np.ctx, "Failed to read netplan config: %v", err)
			return err
		}
		var nc1 netplan.Network
		yaml.Unmarshal(netplanConfig, &nc1)
		if err := mergo.Merge(&nc, nc1); err != nil {
			np.logger.Errorf(np.ctx, "Failed to merge netplan configs: %v", err)
			return err
		}

	}
	for _, file := range files {
		os.Rename(file, file+".bak")
	}
	missing, err := np.FindMissingPhysicalInterfaces(nc)
	if err != nil {
		np.logger.Errorf(np.ctx, "Failed to find missing physical interfaces: %v", err)
		return err
	}
	// 添加不存在的接口
	for _, ifaceName := range missing {
		if ifaceName == "lo" {
			continue
		}
		nc.Network.Ethernets[ifaceName] = &netplan.Ethernet{
			Device: netplan.Device{
				DHCP4:    yamlnillable.BoolOf(true),
				Optional: yamlnillable.BoolOf(true),
			},
		}
	}
	if err := np.SaveNetplanConfig(nc); err != nil {
		os.Remove(configPath)
		for _, file := range files {
			os.Rename(file+".bak", file)
		}
		np.logger.Errorf(np.ctx, "Failed to save netplan config: %v", err)
		return err
	}
	return nil
}
