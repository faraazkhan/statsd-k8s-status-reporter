// Package reporter provides a reporting function for component statuses
package reporter

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getAdditionalTags() []string {
	tags := envOrDefault("STATSD_AGENT_TAGS", "") // Comma separated list of tags for DD Agent
	tagsList := strings.Split(",", tags)
	return tagsList
}

func isKubeDNSUnhealthy() (status bool, err error) {
	hostsString := envOrDefault("STATSD_DNS_LOOKUP_HOSTS", "google.com,kubernetes.svc.default.cluster.local")
	hosts := strings.Split(hostsString, ",")

	for _, host := range hosts {
		_, err = net.LookupHost(host)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func reportToStatsd(status statsd.ServiceCheckStatus, message string) {
	statsdHost := envOrDefault("STATSD_SERVICE_HOST", "statsd-service")
	statsdPort := envOrDefault("STATSD_SERVICE_PORT", "8125")

	statsdURL := fmt.Sprintf("%s:%s", statsdHost, statsdPort)

	statusName := "healthy"

	if status != statsd.Ok {
		statusName = "unhealthy"
	}

	c, err := statsd.New(statsdURL)

	if err != nil {
		log.Printf("There was an error reporting to statsd: %v\n", err)
	}

	c.Namespace = "kubernetes.custom"

	for _, tag := range getAdditionalTags() {
		c.Tags = append(c.Tags, tag)
	}

	hostName := envOrDefault("STATSD_KUBERNETES_CLUSTER_NAME", "minikube")

	check := statsd.ServiceCheck{
		Hostname: hostName,
		Status:   status,
		Tags:     c.Tags,
		Message:  message,
		Name:     envOrDefault("STATSD_CUSTOM_METRIC_NAME", "kubernetes.custom.clusterstatus"),
	}

	err = c.ServiceCheck(&check)

	if err != nil {
		log.Printf("There was an error reporting to statsd at %v : %v\n", statsdURL, err)
	} else {
		log.Printf("Reported status for cluster %v to %v : %v, %v", hostName, statsdURL, statusName, message)
	}
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Monitor important components of the Kubernetes Cluster
func Monitor(clientset *kubernetes.Clientset) {

	pollInterval, err := strconv.Atoi(envOrDefault("STATSD_COMPONENTCHECK_POLL_INTERVAL_SECONDS", "30"))

	if err != nil {
		log.Printf("Could not determine pollInterval: %v\n", err)
		log.Print("Falling back to default pollInterval of 10 seconds")
		pollInterval = 10
	}

	for {
		clusterHealth := statsd.Ok
		message := envOrDefault("STATSD_COMPONENTCHECK_SUCCESS_MESSAGE", "All components are healthy!")
		_, err := isKubeDNSUnhealthy()
		if err != nil {
			message = fmt.Sprintf("There was an error with KubeDNS: %v\n", err)
			clusterHealth = statsd.Critical
		} else {
			components, err := clientset.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
			if err != nil {
				message = fmt.Sprintf("Error listing components: %v\n", err)
				clusterHealth = statsd.Critical
				log.Print(message)
			} else {
				for _, component := range components.Items {
					if component.Conditions[0].Status != "True" {
						clusterHealth = statsd.Critical
						message = fmt.Sprintf("%v is unhealthy: %v", component.Name, component.Conditions[0].Message)
						log.Print(mesearage)
						break // Break on the first unhealthy component
					}
				}
			}
		}
		reportToStatsd(clusterHealth, message)
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}

}
