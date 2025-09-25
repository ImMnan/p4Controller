package k8s

type InitOps struct {
	sts []StsData
}

type DeleteOps struct {
	sts []StsData
}

func syncState(sts []StsData, config Config) (InitOps, DeleteOps, error) {

	// Sync the config map with the current state of the StatefulSets
	// I think we only need to sync the stsName and PodName fields
	// If the item exists in sts and not in config, add it to the DeleteOps
	// If the item exists in config and not in sts, add it to the InitOps
	initOps := InitOps{}
	deleteOps := DeleteOps{}

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
			deleteOps.sts = append(deleteOps.sts, s)
		}
	}

	// Items in config but not in sts -> InitOps
	for key, val := range configSet {
		if _, ok := stsSet[key]; !ok {
			sc := val.Config
			podName := val.PodName
			initOps.sts = append(initOps.sts, StsData{
				StsName:  sc.StsName,
				PodType:  sc.PodType,
				PodName:  podName,
				PodPort:  sc.PodPort,
				Services: sc.Services,
				Init:     sc.InitConfig.Init,
			})
		}
	}

	return initOps, deleteOps, nil
}


