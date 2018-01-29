// main implements the root logic
package main

import (
	"github.com/faraazkhan/statsd-k8s-status-reporter/reporter"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error()) // Only works inside a K8S Cluster
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error()) // Only works inside a K8S Cluster
	}

	reporter.Report(clientset)
}
