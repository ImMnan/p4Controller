package main

import (
	"strconv"
	"strings"

	"github.com/immnan/p4controller/k8s"
	"github.com/immnan/p4controller/p4c"
)

type SyncConfig struct {
	DeleteConfig k8s.Config
	InitConfig   k8s.Config
}

func syncP4Config(config k8s.Config) (SyncConfig, error) {

	// Matches at the Config Struct and ServerJSON.
	// match is done with the name of the server, which is the key in the map and the Name field in the ServerJSON struct.

	// If the item is not found in config, but present in ServerJSON, add it to the InitConfig
	// If the item is found in config, but missing in ServerJSON, add it to the DeleteConfig
	// If the item is found in both, do nothing

	server, err := p4c.ServersRead()
	if err != nil {
		return SyncConfig{}, err
	}

	result := SyncConfig{
		DeleteConfig: k8s.Config{},
		InitConfig:   k8s.Config{},
	}
	result.InitConfig.P4CSpec = make(map[string]k8s.ServerConfig)
	// Build a map of ServerJSON by Name for quick lookup
	itemMap := make(map[string]p4c.ServerJSON)
	for _, srv := range server {
		itemMap[srv.Name] = srv
	}

	// Find servers in p4Servers but not in config.P4CSpec (to InitConfig)
	for name, srv := range itemMap {
		// Extract port from address
		port := 0
		parts := strings.Split(srv.Address, ":")
		if len(parts) == 2 {
			if p, err := strconv.Atoi(parts[1]); err == nil {
				port = p
			}
		}
		if _, exists := config.P4CSpec[name]; !exists {
			result.InitConfig.P4CSpec[name] = k8s.ServerConfig{
				StsName:     srv.Name,
				PodType:     srv.Services,
				PodPort:     port,
				Services:    srv.Services,
				Type:        srv.Type,
				Description: srv.Description,
				Address:     srv.Address,
				InitConfig: k8s.InitConfig{
					Init:        false,
					P4dRootPath: "",
					CtrMounts:   nil,
				},
			}
		}
	}

	// Find servers in config.P4CSpec but not in item (to DeleteConfig)
	for name := range config.P4CSpec {
		// Extract base name (before "-0")
		baseName := name
		if idx := strings.LastIndex(name, "-0"); idx != -1 {
			baseName = name[:idx]
		}
		if _, exists := itemMap[baseName]; !exists {
			if result.DeleteConfig.P4CSpec == nil {
				result.DeleteConfig.P4CSpec = make(map[string]k8s.ServerConfig)
			}
			result.DeleteConfig.P4CSpec[name] = config.P4CSpec[name]
		}
	}

	return result, nil
}
