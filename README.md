# Wavefront Pod Inspector

The *Wavefront Pod Inspector* is an app that uses the Wavefront API to get data on pods in a specific Kubernetes cluster. Using environment variables you can specify which metric and which pod should be looked at.

## Requirements

To be able to build and run the app you'll need:

* Go 1.12 or higher
* A Wavefront API Token (see the [docs](https://docs.wavefront.com/wavefront_api.html) on how to get one)
* Docker (optional, in case you want to run the app as a Docker container)

## Build

To build a standalone executable, run

```bash
export GOPROXY=https://gocenter.io
CGO_ENABLED=0 go build --ldflags "-s -w" -o wavefront .
```

To build a Docker image with the app embedded in it, run

```bash
docker build . -t vmwarecloudadvocacy/wavefront
```

## Run

To run the app, there are four mandatory environment variables that need to be set:

* **METRIC**: metric to query ingested points for (cannot contain wildcards)
* **CLUSTER**: the Kubernetes cluster to look at (cannot contain wildcards)
* **POD_NAME**: the pod name to inspect (cannot contain wildcards)
* **API_TOKEN**: a Wavefront API token

There are two optional environment variables that can be set:

* **TIME_LIMIT**: the duration in time from 'now' the metric data will be requested (must end with a qualifier like `s`, `m`, or `h`. Defaults to `30s`)
* **THRESHOLD**: the threshold setting, above which an alert is printed (defaults to `1`)

To run the app as a standalone executable, using all the above settings:

```bash
export METRIC=heapster.pod.cpu.usage_rate
export CLUSTER=fitcycle-api-dev-k8s-cluster
export POD_NAME=kube-scheduler-ip-172-20-45-146.us-west-2.compute.internal
export API_TOKEN=xyz
export TIME_LIMIT=30s
export THRESHOLD=0.9
./wavefront
```

To run the Docker image and pass in the variables as command-line arguments:

```
docker run --rm -it -e SOURCE=api2-fit-b-m-us-e1-00-m -e METRIC=heapster.pod.cpu.usage_rate -e CLUSTER=fitcycle-api-dev-k8s-cluster -e API_TOKEN=xyz -e POD_NAME=kube-scheduler-ip-172-20-45-146.us-west-2.compute.internal -e TIME_LIMIT=30s -e THRESHOLD=0.9 vmwarecloudadvocacy/wavefront
```

## Output

When you run the app, the output will either be an "alert" (when the average value is above the threshold value) or a message indicating all is okay.

```bash
# Average exceeds threshold
ALERT! avg heapster.pod.cpu.usage_rate: 400.000000

# Average below threshold
No worries, the avg heapster.pod.cpu.usage_rate is 0.802476 (which is less than 0.900000)
```

## Use in GitLab CI

To use the app in GitLab CI as part of a processing step

```bash
if [ -f "alert" ]; then exit 1 && echo "alert"; else echo "Within range. Continuing!"; fi
```
