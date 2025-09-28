package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// Define ClientSet as the Kubernetes clientset type
type ClientSet = kubernetes.Clientset

type InitObj struct {
	Sts []StsData
}

type DeleteObj struct {
	Sts []StsData
}

func SyncConfig(config Config, cs *ClientSet) (InitObj, DeleteObj, error) {

	// Sync the config map with the current state of the StatefulSets
	// I think we only need to sync the stsName and PodName fields
	// If the item exists in sts and not in config, add it to the DeleteOps
	// If the item exists in config and not in sts, add it to the InitOps
	sts, err := stsRead(cs)
	if err != nil {
		return InitObj{}, DeleteObj{}, err
	}

	initObj := InitObj{}
	deleteObj := DeleteObj{}

	// Build sets for quick lookup
	stsSet := make(map[string]StsData) // key: stsName|podName
	for _, s := range sts {
		key := s.StsName + "|" + s.PodName
		stsSet[key] = s
	}

	configSet := make(map[string]struct {
		PodName string
		Config  ServerConfig
	}) // key: stsName|podName
	for podName, sc := range config.P4CSpec {
		key := sc.StsName + "|" + podName
		configSet[key] = struct {
			PodName string
			Config  ServerConfig
		}{PodName: podName, Config: sc}
	}

	// Items in sts but not in config -> DeleteOps
	for key, s := range stsSet {
		// Only add to DeleteOps if there is no config entry with matching StsName and PodName
		if _, ok := configSet[key]; !ok {
			deleteObj.Sts = append(deleteObj.Sts, s)
		}
	}

	// Items in config but not in sts -> InitOps
	for key, val := range configSet {
		if _, ok := stsSet[key]; !ok {
			sc := val.Config
			podName := val.PodName
			initObj.Sts = append(initObj.Sts, StsData{
				StsName:  sc.StsName,
				PodType:  sc.PodType,
				PodName:  podName,
				PodPort:  sc.PodPort,
				Services: sc.Services,
				Init:     sc.InitConfig.Init,
			})
		}
	}

	return initObj, deleteObj, nil
}

func SyncState(initObj InitObj, deleteObj DeleteObj, cs *ClientSet) error {

	for i, sts := range initObj.Sts {
		fmt.Printf("InitOps %d: %+v\n", i, sts)
		err := stsDeployer(cs, []StsData{sts})
		if err != nil {
			return fmt.Errorf("failed to deploy StatefulSet %s: %w", sts.StsName, err)
		}
	}

	for i, sts := range deleteObj.Sts {
		fmt.Printf("DeleteOps %d: %+v\n", i, sts)
		err := stsDeleter(cs, []StsData{sts})
		if err != nil {
			return fmt.Errorf("failed to delete StatefulSet %s: %w", sts.StsName, err)
		}
	}

	return nil
}
