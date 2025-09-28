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
	fmt.Printf("Current StatefulSets in cluster: %+v\n\n", sts)

	initObj := InitObj{}
	deleteObj := DeleteObj{}

	stsSet := make(map[string]StsData) // key: podName
	for _, s := range sts {
		stsSet[s.PodName] = s
	}

	// Items in sts but not in config -> DeleteOps (ignore if podName exists in both)
	for podName, s := range stsSet {
		if _, ok := config.P4CSpec[podName]; !ok {
			deleteObj.Sts = append(deleteObj.Sts, s)
		} else {
			fmt.Println("Pod exists in both config and sts, ignoring: ", podName)
		}
	}

	// Items in config but not in sts -> InitOps (ignore if podName exists in both)
	for podName, sc := range config.P4CSpec {
		if _, ok := stsSet[podName]; !ok {
			initObj.Sts = append(initObj.Sts, StsData{
				StsName:  sc.StsName,
				PodType:  sc.PodType,
				PodName:  podName,
				PodPort:  sc.PodPort,
				Services: sc.Services,
				Init:     sc.InitConfig.Init,
			})
		} else {
			fmt.Println("Pod exists in both config and sts, ignoring: ", podName)
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
