package p4c

import "github.com/immnan/p4controller/k8s"

// this package will initialize the p4d worker servers

func ServerInit(sc k8s.ServerConfig) error {

	if sc.Services == "replica-server" {

		p4ReplicaInit(sc)

	}

	return nil
}

func p4ReplicaInit(sc k8s.ServerConfig) error {
	
}
