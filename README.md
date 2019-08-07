# WaveFront Check App

## Build

You can build the app as a standalone executable (Go app) or as a Docker image

### Go app

```bash
export GOPROXY=https://gocenter.io
CGO_ENABLED=0 go build --ldflags "-s -w" -o wavefront .
```

### Docker image

```bash
docker build . -t vmwarecloudadvocacy/wavefront
```

## Run

To run the app, there are three mandatory environment variables that need to be set:

* SOURCE: source to query ingested points for (cannot contain wildcards). host or source is equivalent, only one should be used.
* METRIC: metric to query ingested points for (cannot contain wildcards)
* API_TOKEN: a Wavefront API token (see the [docs](https://docs.wavefront.com/wavefront_api.html) on how to get one)

There are two optional environment variables that can be set

* TIME_LIMIT: the duration in time from 'now' the metric data will be requested (must end with a qualifier like `s`, `m`, or `h`. Defaults to `30s`)
* THRESHOLD: the threshold setting, above which an alert is printed (defaults to `1`)

### Go app

To run the app as a standalone executable, using all the above settings:

```bash
export SOURCE=api2-fit-b-m-us-e1-00-m
export METRIC=cpu.usage.user
export API_TOKEN=xyz
export TIME_LIMIT=30s
export THRESHOLD=0.9
./wavefront
```

### Docker image

To run the Docker image and pass in the variables as command-line arguments:

```
docker run --rm -it -e SOURCE=api2-fit-b-m-us-e1-00-m -e METRIC=cpu.usage.user -e API_TOKEN=xyz -e TIME_LIMIT=30s -e THRESHOLD=0.9 vmwarecloudadvocacy/wavefront
```

## Output

### With alert

```bash
ALERT! avg CPU usage: 1.100968
```

### No alert

```bash
No worries, the avg CPU usage is 0.802476 (which is less than 0.900000)
```

## Use in GitLab CI

To use the app in GitLab CI as part of a processing step

```bash
if [ -f "alert" ]; then exit 1 && echo "alert"; else echo "Within range. Continuing!"; fi
```
