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
docker build . -t retgits/wavefront-pod-inspector
```

## Run

### Mandatory environment variables

* **GITLAB_TOKEN** the GitLab API token
* **API_TOKEN**: a Wavefront API token
* **POD_NAME**: the pod name to inspect (cannot contain wildcards)
* **CI_PROJECT_NAME** the name of the project in GitLab

### Optional environment variables

* **WAVEFRONT_VARIABLE** the name of the variable in GitLab CI to update with the Wavefront result (defaults to `abc`)
* **METRIC**: metric to query ingested points for (cannot contain wildcards, defaults to `kubernetes.pod_container.cpu.usage_rate`)
* **CLUSTER**: the Kubernetes cluster to look at (cannot contain wildcards, defaults to `acmefitness-aks-02`)
* **THRESHOLD**: the threshold setting, above which an alert is printed (defaults to `1`)

To run the app as a standalone executable, using all the above settings:

```bash
export GITLAB_TOKEN=def
export API_TOKEN=xyz
export POD_NAME=kube-scheduler-ip-172-20-45-146.us-west-2.compute.internal
export CI_PROJECT_NAME=vmworld2019-tim
export WAVEFRONT_VARIABLE=abc
export METRIC=kubernetes.pod_container.cpu.usage_rate
export CLUSTER=acmefitness-aks-02
export THRESHOLD=0.9
./wavefront
```

With just the mandatory environment variables:

```bash
export GITLAB_TOKEN=def
export API_TOKEN=xyz
export POD_NAME=kube-scheduler-ip-172-20-45-146.us-west-2.compute.internal
export CI_PROJECT_NAME=vmworld2019-tim
./wavefront
```

## Output

When you run the app, the output will either be an "alert" (when the average value is above the threshold value) or a message indicating all is okay.

```bash
--- Configuration Settings ---
Wavefront Variable: abc
Metric            : kubernetes.pod_container.cpu.usage_rate
Cluster           : acmefitness-aks-02
Pod               : tunnelfront-54989596f-t274l
Threshold         : 1.000000
GitLab Project    : vmworld2019-tim

---  Calling Wavefront on  ---
https://try.wavefront.com/api/v2/chart/api?cached=true&g=h&q=ts%28%22kubernetes.pod_container.cpu.usage_rate%22%2C+cluster%3D%22acmefitness-aks-02%22+and+pod_name%3D%22tunnelfront-54989596f-t274l%22%29&s=1565980438617&sorted=false&strict=true
Wavefront response: 200 OK

---   Calling GitLab on    ---
https://gitlab.com/api/v4/projects/vmware-cloud-advocacy%2fvmworld2019-tim/variables/abc
Setting abc to failed <-- this will be either failed or passed depending on whether the value is below the threshold or not
GitLab response: 200 OK

--- Wavefront Check Result ---
ALERT! avg kubernetes.pod_container.cpu.usage_rate: 72.029412% <-- this message will show 'No worries' when the value is below the threshold
```

When the output is above the threshold (when the alert text is displayed), the app also tries to update an environment variable setting in GitLab CI

## Use in GitLab CI

To use the app in GitLab CI as part of a processing step

Either use the textfile which is created

```bash
if [ -f "alert" ]; then exit 1 && echo "alert"; else echo "Within range. Continuing!"; fi
```

Or use the environment variable that was passed in.
