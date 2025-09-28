package main

import (
	"time"

	"github.com/immnan/p4controller/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// We will have a function that will manage 2 packages through 2 different channels CH1 and CH2
// Both the channel will send and receive data from each other.

// CH1 will run the k8s package functions
// CH2 will run the p4c package functions

//

func getClientSet() *kubernetes.Clientset {
	// Create a new Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//return clientset
	return clientset

}

func p4Controller() {
	cs := getClientSet()
	err := k8s.ConfigMapInit(cs)
	if err != nil {
		panic(err)
	}

	chK8sToP4c := make(chan k8s.Config, 2)
	chP4cToK8s := make(chan SyncConfig, 2)

	// CH1: k8s package functions
	go func() {
		for {
			k8sConfig, _, err := k8s.ConfigReader(cs)
			if err != nil {
				panic(err)
			}
			chK8sToP4c <- k8sConfig // Send to CH2

			SyncConfig := <-chP4cToK8s // Receive from CH2

			if !SyncConfig.IsEmpty() {
				if !IsK8sConfigEmpty(SyncConfig.InitConfig) {
					if err := k8s.ConfigWriter(cs, SyncConfig.InitConfig); err != nil {
						panic(err)
					}
				}
				if !IsK8sConfigEmpty(SyncConfig.DeleteConfig) {
					if err := k8s.ConfigDeleter(cs, SyncConfig.DeleteConfig); err != nil {
						panic(err)
					}
				}
			}

			time.Sleep(120 * time.Second)
		}
	}()

	// CH2: p4c package functions
	go func() {
		for {
			k8sConfig := <-chK8sToP4c // Receive from CH1
			SyncConfig, err := syncP4Config(k8sConfig)
			if err != nil {
				panic(err)
			}
			chP4cToK8s <- SyncConfig // Send to CH1
			time.Sleep(300 * time.Second)
		}
	}()

	select {} // Block forever
}

// Add this for robust empty checks
func IsK8sConfigEmpty(c k8s.Config) bool {
	if len(c.P4CSpec) > 0 {
		return false
	}
	return true
}

func (sc SyncConfig) IsEmpty() bool {
	return IsK8sConfigEmpty(sc.InitConfig) && IsK8sConfigEmpty(sc.DeleteConfig)
}
