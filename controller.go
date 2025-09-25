package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// We will have a function that will manage 2 packages through 2 different channels CH1 and CH2
// Both the channel will send and receive data from each other.

// CH1 will run the k8s package functions
// CH2 will run the p4c package functions

type ClientSet struct {
	clientset *kubernetes.Clientset
}

func (cs *ClientSet) getClientSet() {
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
	cs.clientset = clientset

}

func channelManager() {

	// Create 2 channels
	ch1 := make(chan string)
	ch2 := make(chan string)

	// Create a new ClientSet
	cs := &ClientSet{}
	cs.getClientSet()

	k8s.configReader(cs)

}
