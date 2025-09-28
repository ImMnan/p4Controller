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
	sts, err := stsRead(cs)
	if err != nil {
		return InitObj{}, DeleteObj{}, err
	}

	initObj := InitObj{}
	deleteObj := DeleteObj{}

	// Build sets for quick lookup using podName and stsName as the key
	stsSet := make(map[string]StsData) // key: stsName|podName
	podNamesInSts := make(map[string]struct{})
	for _, s := range sts {
		key := s.StsName + "|" + s.PodName
		stsSet[key] = s
		podNamesInSts[s.PodName] = struct{}{}
	}

	configSet := make(map[string]ServerConfig) // key: stsName|podName
	for podName, sc := range config.P4CSpec {
		key := sc.StsName + "|" + podName
		configSet[key] = sc
	}

	// Items in sts but not in config -> DeleteOps, unless podName exists in config
	for key, s := range stsSet {
		if _, ok := configSet[key]; !ok {
			if _, exists := config.P4CSpec[s.PodName]; exists {
				// PodName exists in config, so ignore (desired state)
				continue
			}
			deleteObj.Sts = append(deleteObj.Sts, s)
		}
	}

	// Items in config but not in sts -> InitOps, unless podName exists in sts
	for key, sc := range configSet {
		if _, ok := stsSet[key]; !ok {
			if _, exists := podNamesInSts[key[len(sc.StsName)+1:]]; exists {
				// PodName exists in sts, so ignore (desired state)
				continue
			}
			initObj.Sts = append(initObj.Sts, StsData{
				StsName:  sc.StsName,
				PodType:  sc.PodType,
				PodName:  key[len(sc.StsName)+1:], // extract podName from key
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
