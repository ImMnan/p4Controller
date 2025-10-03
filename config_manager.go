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
	server, err := p4c.ServersRead()
	if err != nil {
		return SyncConfig{}, err
	}

	result := SyncConfig{
		DeleteConfig: k8s.Config{},
		InitConfig:   k8s.Config{},
	}
	result.InitConfig.P4CSpec = make(map[string]k8s.ServerConfig)

	// Build a map of ServerJSON by Name (StatefulSet name) for quick lookup
	itemMap := make(map[string]p4c.ServerJSON)
	for _, srv := range server {
		itemMap[srv.Name] = srv
	}

	// Find servers in ServerJSON but not in config.P4CSpec (to InitConfig)
	for stsName, srv := range itemMap {
		// Check if any config entry matches this stsName
		found := false
		for _, sc := range config.P4CSpec {
			if sc.StsName == stsName {
				found = true
				break

			}
		}
		if !found {
			// Extract port from address
			port := 0
			parts := strings.Split(srv.Address, ":")
			if len(parts) == 2 {
				if p, err := strconv.Atoi(parts[1]); err == nil {
					port = p
				}
			}
			// Use stsName as key for InitConfig
			result.InitConfig.P4CSpec[stsName] = k8s.ServerConfig{
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

	// Find servers in config.P4CSpec but not in ServerJSON (to DeleteConfig)
	for podName, sc := range config.P4CSpec {
		stsName := sc.StsName
		if _, exists := itemMap[stsName]; !exists {
			if result.DeleteConfig.P4CSpec == nil {
				result.DeleteConfig.P4CSpec = make(map[string]k8s.ServerConfig)
			}
			result.DeleteConfig.P4CSpec[podName] = sc
		}
	}

	return result, nil
}
