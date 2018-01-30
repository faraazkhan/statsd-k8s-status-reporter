// Package reporter provides a reporting function for component statuses
package reporter

import (
	"fmt"
	"log"
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
		log.Printf("Reported status for cluster %v to %v : %v", hostName, statsdURL, statusName)
	}
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Report metrics to a Statsd Collector
func Report(clientset *kubernetes.Clientset) {

	clusterHealth := statsd.Ok
	message := envOrDefault("STATSD_COMPONENTCHECK_SUCCESS_MESSAGE", "All components are healthy!")
	pollInterval, err := strconv.Atoi(envOrDefault("STATSD_COMPONENTCHECK_POLL_INTERVAL_SECONDS", "10"))

	if err != nil {
		log.Printf("Could not determine pollInterval: %v\n", err)
		log.Printf("Falling back to default pollInterval of 10 seconds")
		pollInterval = 10
	}

	for {
		components, err := clientset.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
		if err != nil {
			message = fmt.Sprintf("Error listing components: %v\n", err)
			log.Printf(message)
			reportToStatsd(statsd.Critical, message) // Assume we are having trouble hitting the API, Currently this will misreport for RBAC issues

		} else {
			for idx, component := range components.Items {
				if component.Conditions[0].Status != "True" {
					clusterHealth = statsd.Critical
					message = fmt.Sprintf("%v is unhealthy: %v", component.Name, component.Conditions[0].Message)
					log.Printf(message)
					reportToStatsd(clusterHealth, message)
					break // Break on the first unhealthy component
				}
				if idx+1 == len(components.Items) {
					reportToStatsd(clusterHealth, message) // We did not find an unhealthy component from the entire components.Item list
				}
			}
		}
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}

}
