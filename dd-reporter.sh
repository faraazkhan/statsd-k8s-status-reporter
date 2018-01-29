#!/usr/bin/env bash

cat <<EOF
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: statsd-k8s-status-reporter
  namespace: default
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: statsd-k8s-status-reporter
rules:
  - apiGroups: [""]
    resources:
      - componentstatuses
    verbs: ["list"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: statsd-k8s-status-reporter
subjects:
  - kind: ServiceAccount
    name: statsd-k8s-status-reporter
    namespace: default
roleRef:
  kind: ClusterRole
  name: statsd-k8s-status-reporter
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: statsd-k8s-status-reporter
  labels:
    app: statsd-k8s-status-reporter
spec:
  selector:
    matchLabels:
      app: statsd-k8s-status-reporter
  replicas: 1
  template:
    metadata:
      labels:
        app: statsd-k8s-status-reporter
    spec:
      containers:
        - name: reporter
          image: ${REPORTER_IMAGE}:${REPORTER_TAG}
          imagePullPolicy: IfNotPresent
          env:
          - name: STATSD_KUBERNETES_CLUSTER_NAME
            value: mycluster.rationalizeit.us # change me
          - name: STATSD_SERVICE_HOST
            value: dd-statsd-service # change me
      serviceAccountName: statsd-k8s-status-reporter
EOF
