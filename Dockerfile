FROM golang:alpine
WORKDIR /go/src/github.com/faraazkhan/statsd-k8s-status-reporter
COPY . .
RUN GOOS=linux go build -o ./app .

FROM alpine:latest
COPY --from=0 /go/src/github.com/faraazkhan/statsd-k8s-status-reporter/app /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
