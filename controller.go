package main

import (
	"fmt"
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
		fmt.Println("ConfigMapInit error:", err)
	}

	chK8sToP4c := make(chan k8s.Config, 2)
	chP4cToK8s := make(chan SyncConfig, 2)

	// CH1: k8s package functions
	go func() {
		for {
			k8sConfig, _, err := k8s.ConfigReader(cs)
			if err != nil {
				fmt.Println("ConfigReader error:", err)
				time.Sleep(120 * time.Second)
				continue
			}
			initObj, delObj, err := k8s.SyncConfig(k8sConfig, cs)
			if err != nil {
				fmt.Println("SyncConfig error:", err)
				time.Sleep(120 * time.Second)
				continue
			}

			if err := k8s.SyncState(initObj, delObj, cs); err != nil {
				fmt.Println("SyncState error:", err)
				time.Sleep(120 * time.Second)
				continue
			}

			chK8sToP4c <- k8sConfig // Send to CH2

			SyncConfig := <-chP4cToK8s // Receive from CH2

			if !SyncConfig.IsEmpty() {
				if !IsK8sConfigEmpty(SyncConfig.InitConfig) {
					if err := k8s.ConfigWriter(cs, SyncConfig.InitConfig); err != nil {
						fmt.Println("ConfigWriter error:", err)
					}
				}
				if !IsK8sConfigEmpty(SyncConfig.DeleteConfig) {
					if err := k8s.ConfigDeleter(cs, SyncConfig.DeleteConfig); err != nil {
						fmt.Println("ConfigDeleter error:", err)
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
				fmt.Println("syncP4Config error:", err)
				chP4cToK8s <- SyncConfig // Send empty SyncConfig on error
			} else {
				chP4cToK8s <- SyncConfig // Send to CH1
			}
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
