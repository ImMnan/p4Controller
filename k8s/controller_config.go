package k8s

import (
	"context"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	P4CSpec map[string]ServerConfig `yaml:"p4CSpec"`
}

type ServerConfig struct {
	StsName     string     `yaml:"stsName"`  // StatefulSet name
	PodType     string     `yaml:"podType"`  // master or worker
	PodPort     int        `yaml:"podPort"`  // Port on which p4d server will run inside the pod
	Services    string     `yaml:"services"` // commit-server, edge-server, etc
	Type        string     `yaml:"type"`     // server, proxy, connector etc
	Address     string     `yaml:"address"`  // P4C:Port
	Description string     `yaml:"description"`
	InitConfig  InitConfig `yaml:"initConfig"`
	SvcIPPort   string     `yaml:"svcIPPort"`
	// DistributedConfig string     `yaml:"distributedConfig"`
}

type InitConfig struct {
	Init        bool       `yaml:"init"`
	P4dRootPath string     `yaml:"p4dRootPath"`
	CtrMounts   []CtrMount `yaml:"ctrMounts"`
}

type CtrMount struct {
	MountPath string `yaml:"mountPath"`
	PVC       string `yaml:"pvc"`
	PV        string `yaml:"pv"`
}

// This package function will be responsible for configuring the config map of the controller
// READ, WRITE & SYNC

func ConfigMapInit(cs *ClientSet) error {
	// Initialize the config map
	configMapName := os.Getenv("CONTROLLER_CM_NAME")
	if configMapName == "" {
		configMapName = "p4controller-cm"
	}
	ns := os.Getenv("WORKING_NAMESPACE")
	if ns == "" {
		ns = "default"
	}
	_, err := cs.CoreV1().ConfigMaps(ns).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		// If the config map doesn't exist, create it
		fmt.Println("configMap not found, exiting")
		return err
	}
	// If the config map exists, read it
	// exit the function successfully
	fmt.Println("configMap found.")
	return nil
}

func ConfigReader(cs *ClientSet) (Config, *v1.ConfigMap, error) {
	var config Config
	ns := os.Getenv("WORKING_NAMESPACE")
	if ns == "" {
		ns = "default"
	}
	configMapName := os.Getenv("CONTROLLER_CM_NAME")
	if configMapName == "" {
		configMapName = "p4controller-cm"
	}
	// Get the existing config map
	cm, err := cs.CoreV1().ConfigMaps(ns).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Error getting config map:", err)
		return Config{}, nil, err
	}
	content, ok := cm.Data["p4Controller_config.yaml"]
	if !ok {
		fmt.Println("Config file not found in config map")
		return Config{}, nil, fmt.Errorf("config file not found in config map")
	}
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		fmt.Println("Error parsing config file:", err)
		return Config{}, nil, err
	}
	return config, cm, nil
}

func ConfigWriter(cs *ClientSet, config Config) error {
	// Read the existing config and configmap
	cfg, cm, err := ConfigReader(cs)
	if err != nil {
		fmt.Println("Error reading existing config:", err)
		return err
	}
	ns := os.Getenv("WORKING_NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	// Merge new entries into existing config
	if cfg.P4CSpec == nil {
		cfg.P4CSpec = make(map[string]ServerConfig)
	}
	for key, val := range config.P4CSpec {
		cfg.P4CSpec[key] = val // Add or update entry
	}

	// Marshal the merged config struct to YAML
	yamlData, err := yaml.Marshal(&cfg)
	if err != nil {
		fmt.Println("Error marshaling config to YAML:", err)
		return err
	}

	// Update the data field
	cm.Data["p4Controller_config.yaml"] = string(yamlData)

	// Update the config map in Kubernetes
	_, err = cs.CoreV1().ConfigMaps(ns).Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println("Error updating config map:", err)
		return err
	}

	fmt.Println("Config map updated successfully.")
	return nil
}

func ConfigDeleter(cs *ClientSet, config Config) error {
	ns := os.Getenv("WORKING_NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	cfg, cm, err := ConfigReader(cs)
	if err != nil {
		fmt.Println("Error reading existing config:", err)
		return err
	}

	content, ok := cm.Data["p4Controller_config.yaml"]
	if !ok {
		return fmt.Errorf("config file not found in config map")
	}

	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		fmt.Println("Error parsing config file:", err)
		return err
	}

	// Delete the dictionary key from the data field
	for key := range config.P4CSpec {
		delete(cfg.P4CSpec, key)
	}

	// Marshal back to YAML
	yamlData, err := yaml.Marshal(&cfg)
	if err != nil {
		fmt.Println("Error marshaling config to YAML:", err)
		return err
	}

	// Update the config map data
	cm.Data["p4Controller_config.yaml"] = string(yamlData)

	// Update the config map in Kubernetes
	_, err = cs.CoreV1().ConfigMaps(ns).Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println("Error updating config map:", err)
		return err
	}

	fmt.Println("Config entry deleted successfully.")
	return nil
}

//func configFinalize(config Config, cs *ClientSet) Config {
//	// Rearrange the config map entries if needed
//	// Currently, this function does nothing and returns the config as is
//	ns := os.Getenv("WORKING_NAMESPACE")
//	if ns == "" {
//		ns = "default"
//	}
//	// List StatefulSets with label selectors
//
//	labelSelector := "app=p4d,managed-by=p4controller"
//	stss, err := cs.AppsV1().StatefulSets(ns).List(context.Background(), metav1.ListOptions{
//		LabelSelector: labelSelector,
//	})
//	if err != nil {
//		return Config{}, fmt.Errorf("failed to list StatefulSets: %w", err)
//	}
//	return config
//
//	//
//
//}
