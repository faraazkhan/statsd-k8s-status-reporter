# Kubernetes Component Status Reporter for Statsd

Simple utility pod that reports a summarized kubernetes component status to a statsd collector.

# Why?

Kube State Metrics does not support component statuses because lack of a watcher API
https://github.com/kubernetes/kube-state-metrics/pull/317

# Installation

Before installing the reporter, you must (if you don't already have it) create a statsd service

You can do that (for a standard datadog deployment) by using the provided `dd-service.yaml` file

Simply:

```
kubectl apply -f dd-service.yaml
```

This is distributed as a docker image. To run it inside your k8s cluster you can simply in the same namespace as your datadog agent:

```
kubectl run statsd-k8s-status-reporter --image=faraazkhan/statsd-k8s-status-reporter
```

If you'd rather build your own image (recommended) then clone this project and run:

```
make build
```

See `dd-reporter.sh` for a sample Deployment config or simply change the Repo Name in the `Makefile`
and run `make deploy`

# RBAC

By default the reporter runs in its own service account with limited permissions. If you are creating your own you need:

```
  - apiGroups: [""]
    resources:
      - componentstatuses
    verbs: ["list"]

```

# Configure

The project supports a few env vars:

```
STATSD_KUBERNETES_CLUSTER_NAME is the name your K8s cluster. This is added as a tag on the Service Check in DD.
The default is "minikube"
For example: STATSD_KUBERNETES_CLUSTER_NAME="k8s.myorganization.internal.net"

STATSD_AGENT_TAGS is a comma separated list of tags you want to add to service check for your cluster.
For example: STATSD_AGENT_TAGS="mycostcenter,myawsaccountalias"

STATSD_SERVICE_HOST is the host name of the statsd service you want to push the service checks to. Default: statsd-service

STATSD_SERVICE_PORT is the port of the statsd service you want to push to service checks to. Default: 8125

STATSD_COMPONENTCHECK_SUCCESS_MESSAGE is a string message sent to Datadog when all components are healthy
For example: STATSD_COMPONENTCHECK_SUCCESS_MESSAGE="All components are healthy!"

STATSD_COMPONENTCHECK_POLL_INTERVAL_SECONDS is the number of seconds between each check. Default is 10
Change this if you would like with:
STATSD_COMPONENTCHECK_POLL_INTERVAL_SECONDS="30"
```

A sample datadog monitor is also provided in json format at `sample_dd_monitor.json`
