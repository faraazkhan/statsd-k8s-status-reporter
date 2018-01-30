REPORTER_IMAGE := faraazkhan/statsd-k8s-status-reporter
REPORTER_TAG := $(shell git rev-parse --short HEAD)
REPORTER_NAMESPACE := default

export REPORTER_IMAGE REPORTER_TAG REPORTER_NAMESPACE

all: build deploy

build:
	@glide up
	@docker build -t $(REPORTER_IMAGE):$(REPORTER_TAG) .

delete:
	@kubectl delete -f dd-reporter.yaml

deploy:
	./dd-reporter.sh | kubectl apply -f -
